use std::env;
use std::path::{Path, PathBuf};
use std::process::Command;

fn main() {
    println!("cargo:rerun-if-env-changed=AR");
    println!("cargo:rerun-if-env-changed=CC");
    println!("cargo:rerun-if-changed=build.rs");
    println!("cargo:rerun-if-changed=c/wrapper.c");
    println!("cargo:rerun-if-changed=include/wrapper.h");

    let manifest_dir = PathBuf::from(env::var("CARGO_MANIFEST_DIR").expect("missing manifest dir"));
    let out_dir = PathBuf::from(env::var("OUT_DIR").expect("missing OUT_DIR"));
    let target_os = env::var("CARGO_CFG_TARGET_OS").expect("missing target os");
    let pkg_name = match target_os.as_str() {
        "linux" => "fuse3",
        "macos" => "fuse-t",
        other => panic!("unsupported target OS: {other}"),
    };

    let cflags = pkg_config(["--cflags", pkg_name]);
    let libs = pkg_config(["--libs", pkg_name]);

    let cc = env::var("CC").unwrap_or_else(|_| "clang".to_string());
    let ar = env::var("AR").unwrap_or_else(|_| "ar".to_string());

    let source = manifest_dir.join("c").join("wrapper.c");
    let obj = out_dir.join("wrapper.o");
    let archive = out_dir.join("libfuse_lowlevel_sys_wrapper.a");

    compile_wrapper(&cc, &manifest_dir, &source, &obj, &target_os, &cflags);
    archive_wrapper(&ar, &archive, &obj);

    println!("cargo:rustc-link-search=native={}", out_dir.display());
    println!("cargo:rustc-link-lib=static=fuse_lowlevel_sys_wrapper");

    emit_link_directives(&libs);
}

fn pkg_config<const N: usize>(args: [&str; N]) -> String {
    let output = Command::new("pkg-config")
        .args(args)
        .output()
        .unwrap_or_else(|err| panic!("failed to run pkg-config: {err}"));

    if !output.status.success() {
        panic!(
            "pkg-config {:?} failed:\nstdout: {}\nstderr: {}",
            args,
            String::from_utf8_lossy(&output.stdout),
            String::from_utf8_lossy(&output.stderr)
        );
    }

    String::from_utf8(output.stdout)
        .expect("pkg-config output was not valid UTF-8")
        .trim()
        .to_string()
}

fn compile_wrapper(
    cc: &str,
    manifest_dir: &Path,
    source: &Path,
    obj: &Path,
    target_os: &str,
    cflags: &str,
) {
    let mut command = Command::new(cc);
    command
        .arg("-c")
        .arg("-std=c11")
        .arg("-O2")
        .arg("-Wall")
        .arg("-Wextra")
        .arg("-o")
        .arg(obj)
        .arg(source)
        .arg("-I")
        .arg(manifest_dir.join("include"));

    match target_os {
        "linux" => {
            command.arg("-DFUSELOWLEVEL_TARGET_LINUX=1");
        }
        "macos" => {
            command.arg("-DFUSELOWLEVEL_TARGET_MACOS=1");
            command.arg("-I/Library/Frameworks/fuse_t.framework/Headers");
        }
        _ => unreachable!(),
    }

    for token in cflags.split_whitespace() {
        command.arg(token);
    }

    let status = command
        .status()
        .unwrap_or_else(|err| panic!("failed to invoke compiler {cc}: {err}"));
    if !status.success() {
        panic!("compiling wrapper.c failed with status {status}");
    }
}

fn archive_wrapper(ar: &str, archive: &Path, obj: &Path) {
    let status = Command::new(ar)
        .args([
            "crus",
            archive.to_str().expect("archive path"),
            obj.to_str().expect("obj path"),
        ])
        .status()
        .unwrap_or_else(|err| panic!("failed to invoke archiver {ar}: {err}"));

    if !status.success() {
        panic!("archiving wrapper object failed with status {status}");
    }
}

fn emit_link_directives(libs: &str) {
    for token in libs.split_whitespace() {
        if let Some(path) = token.strip_prefix("-L") {
            println!("cargo:rustc-link-search=native={path}");
            continue;
        }
        if let Some(name) = token.strip_prefix("-l") {
            println!("cargo:rustc-link-lib={name}");
            continue;
        }
        if token.starts_with("-Wl,") {
            println!("cargo:rustc-link-arg={token}");
        }
    }
}
