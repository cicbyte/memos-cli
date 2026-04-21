#!/usr/bin/env python3
"""本地构建脚本 — 交叉编译 memos-cli 并可选 UPX 压缩"""

import os
import shutil
import subprocess
import sys
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
VERSION_FILE = ROOT / "VERSION"
DIST_DIR = ROOT / "dist"

LDFLAGS_PREFIX = "github.com/cicbyte/memos-cli/cmd/version"


def read_version() -> str:
    return VERSION_FILE.read_text().strip()


def run_command(cmd, cwd=None, direct_output=False):
    """运行命令并实时显示输出"""
    if direct_output:
        process = subprocess.Popen(
            cmd, stdout=sys.stdout, stderr=sys.stderr, cwd=cwd,
            universal_newlines=True
        )
        return process.wait()
    else:
        process = subprocess.Popen(
            cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
            text=True, encoding="utf-8", errors="replace", cwd=cwd,
            bufsize=1, universal_newlines=True
        )
        while True:
            output = process.stdout.readline()
            if output == "" and process.poll() is not None:
                break
            if output:
                print(f"  {output.strip()}")
        stderr = process.stderr.read()
        if stderr:
            print(f"  [stderr] {stderr.strip()}")
        return process.poll()


def get_build_info():
    """获取 Git 构建信息"""
    try:
        commit = subprocess.check_output(
            ["git", "rev-parse", "HEAD"], stderr=subprocess.DEVNULL, universal_newlines=True
        ).strip()
    except subprocess.CalledProcessError:
        commit = "unknown"
    build_time = time.strftime("%Y-%m-%dT%H:%M:%S")
    return commit, build_time


def check_upx():
    try:
        return subprocess.run(["upx", "--version"], capture_output=True).returncode == 0
    except FileNotFoundError:
        return False


def compress_with_upx(filepath):
    if not check_upx():
        print("  UPX 未安装，跳过压缩")
        return
    original = os.path.getsize(filepath)
    print(f"  UPX 压缩中... ({original / 1024 / 1024:.2f} MB)")
    ret = run_command(["upx", "--best", filepath], direct_output=True)
    if ret == 0:
        compressed = os.path.getsize(filepath)
        ratio = (1 - compressed / original) * 100
        print(f"  压缩完成: {original / 1024 / 1024:.2f} MB → {compressed / 1024 / 1024:.2f} MB ({ratio:.1f}%)")
    else:
        print("  UPX 压缩失败，跳过")


def build_target(goos, goarch, version, commit, build_time):
    """编译单个目标平台"""
    ext = ".exe" if goos == "windows" else ""
    output_name = f"memos-cli_{goos}_{goarch}{ext}"
    output_path = DIST_DIR / output_name

    ldflags = (
        f"-s -w "
        f"-X {LDFLAGS_PREFIX}.Version={version} "
        f"-X {LDFLAGS_PREFIX}.GitCommit={commit} "
        f"-X {LDFLAGS_PREFIX}.BuildTime={build_time}"
    )

    env = os.environ.copy()
    env["CGO_ENABLED"] = "0"
    env["GOOS"] = goos
    env["GOARCH"] = goarch

    print(f"  编译 {goos}/{goarch}...")
    ret = subprocess.run(
        ["go", "build", "-ldflags", ldflags, "-o", str(output_path), "."],
        cwd=ROOT, env=env
    )
    if ret.returncode != 0:
        print(f"  编译 {goos}/{goarch} 失败!")
        return False
    print(f"  ✓ {output_name} ({os.path.getsize(output_path) / 1024 / 1024:.2f} MB)")

    # 仅对当前平台执行 UPX 压缩
    current_goos = "windows" if os.name == "nt" else "linux" if os.name == "posix" else "darwin"
    if goos == current_goos:
        compress_with_upx(str(output_path))

    return True


def main():
    import argparse
    parser = argparse.ArgumentParser(description="memos-cli 本地构建脚本")
    parser.add_argument("--platform", choices=["windows", "linux", "darwin"], help="仅编译指定平台")
    parser.add_argument("--local", action="store_true", help="仅编译当前平台")
    args = parser.parse_args()

    version = read_version()
    commit, build_time = get_build_info()

    print(f"memos-cli {version} | commit: {commit[:8]} | {build_time}")
    print()

    if DIST_DIR.exists():
        shutil.rmtree(DIST_DIR)
    DIST_DIR.mkdir()

    targets = [
        ("windows", "amd64"),
        ("linux", "amd64"),
        ("darwin", "amd64"),
    ]

    if args.local:
        current = "windows" if os.name == "nt" else "linux" if os.name == "posix" else "darwin"
        targets = [(current, "amd64")]
    elif args.platform:
        targets = [(args.platform, "amd64")]

    success = True
    for goos, goarch in targets:
        if not build_target(goos, goarch, version, commit, build_time):
            success = False

    print()
    if success:
        print(f"构建完成! 输出目录: {DIST_DIR}")
    else:
        print("部分平台构建失败!")
        sys.exit(1)


if __name__ == "__main__":
    main()
