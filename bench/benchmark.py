#!/usr/bin/env python3
"""Benchmark screenfetch, neofetch, fastfetch, hifetch, macchina, and dawnfetch with readable output.

Usage examples:
  python3 bench/benchmark.py
  python3 bench/benchmark.py -n 50 -w 5
  python3 bench/benchmark.py --preset quick
  python3 bench/benchmark.py --tf './dawnfetch --theme transfem'
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import math
import os
import shlex
import shutil
import statistics
import subprocess
import sys
import tempfile
from dataclasses import dataclass
from pathlib import Path
from typing import List


@dataclass
class BenchResult:
    name: str
    runs: int
    mean_ms: float
    median_ms: float
    p95_ms: float
    min_ms: float
    max_ms: float
    stddev_ms: float


class C:
    RESET = "\033[0m"
    BOLD = "\033[1m"
    DIM = "\033[2m"
    RED = "\033[91m"
    GREEN = "\033[92m"
    YELLOW = "\033[93m"
    BLUE = "\033[94m"
    CYAN = "\033[96m"


def color_enabled(force: bool, disable: bool) -> bool:
    if disable:
        return False
    if force:
        return True
    return sys.stdout.isatty() and os.getenv("NO_COLOR") is None


def ctext(text: str, code: str, enable: bool) -> str:
    if not enable:
        return text
    return f"{code}{text}{C.RESET}"


def clickable_path(path: Path) -> str:
    """return a clickable file hyperlink for terminals that support osc 8."""
    resolved = path.resolve()
    plain = str(resolved)
    if not sys.stdout.isatty() or os.getenv("TERM", "").lower() == "dumb":
        return plain
    uri = resolved.as_uri()
    return f"\033]8;;{uri}\033\\{plain}\033]8;;\033\\"


def parse_args() -> argparse.Namespace:
    terms = """terminology:
    warmup:      throwaway runs before timing starts that warm up the system and are never counted.
    runs:        how many times each tool was timed.
    mean(ms):    average runtime across all runs, though a few slow runs can pull it higher.
    median(ms):  the middle runtime when all runs are sorted from fastest to slowest, so extreme values on either end don't distort it.
    p95(ms):     95 out of 100 runs finished faster than this value, making it a good indicator of how slow things get in practice.
    min/max(ms): the single fastest and slowest runs.
    stddev(ms):  measures how consistent the runs were; a lower value means more stable timings.
    """
    parser = argparse.ArgumentParser(
        description="Benchmark different tools for fetching sysinfo",
        epilog=terms,
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument(
        "--preset",
        choices=("quick", "normal", "deep"),
        default="normal",
        help="quick=20/2, normal=40/3, deep=100/5 (runs/warmup)",
    )
    parser.add_argument(
        "-n", "--runs", type=int, default=None, help="timed runs per tool (overrides preset)"
    )
    parser.add_argument(
        "-w", "--warmup", type=int, default=None, help="warmup runs per tool (overrides preset)"
    )
    parser.add_argument(
        "--dawnfetch-cmd",
        "--tf",
        dest="dawnfetch_cmd",
        default="",
        help="explicit dawnfetch command (example: './dawnfetch --no-color')",
    )
    parser.add_argument(
        "--allow-small-sample",
        action="store_true",
        help="allow unreliable tiny samples (runs < 10 or warmup < 1)",
    )
    parser.add_argument(
        "--no-color",
        action="store_true",
        help="disable colored benchmark output",
    )
    parser.add_argument(
        "--force-color",
        action="store_true",
        help="force colored benchmark output",
    )
    parser.add_argument(
        "--plot",
        choices=("on", "off"),
        default="off",
        help="render a horizontal benchmark graph with matplotlib (default: off)",
    )
    return parser.parse_args()


def runs_warmup_from_args(args: argparse.Namespace) -> tuple[int, int]:
    preset_map = {
        "quick": (10, 2),
        "normal": (30, 3),
        "deep": (60, 5),
    }
    runs, warmup = preset_map[args.preset]
    if args.runs is not None:
        runs = args.runs
    if args.warmup is not None:
        warmup = args.warmup
    return runs, warmup


def percentile(sorted_values: List[float], p: float) -> float:
    if not sorted_values:
        return 0.0
    if p <= 0:
        return sorted_values[0]
    if p >= 100:
        return sorted_values[-1]
    idx = int(math.ceil((p / 100.0) * len(sorted_values))) - 1
    idx = max(0, min(idx, len(sorted_values) - 1))
    return sorted_values[idx]


def prepare_dawnfetch_cmd(repo_root: Path, explicit_cmd: str) -> List[str]:
    if explicit_cmd.strip():
        return shlex.split(explicit_cmd)

    local_bin = repo_root / "dawnfetch"
    if local_bin.exists() and os.access(local_bin, os.X_OK):
        return [str(local_bin), "--no-color"]

    if shutil.which("go"):
        tmp_dir = Path(tempfile.mkdtemp(prefix="dawnfetch-bench-"))
        built = tmp_dir / "dawnfetch"
        build_cmd = ["go", "build", "-o", str(built), "."]
        proc = subprocess.run(build_cmd, cwd=repo_root, capture_output=True, text=True)
        if proc.returncode == 0 and built.exists():
            return [str(built), "--no-color"]

    raise RuntimeError(
        "Unable to prepare dawnfetch command. Provide --dawnfetch-cmd or install Go/build ./dawnfetch."
    )


def find_cmd(name: str) -> List[str] | None:
    path = shutil.which(name)
    if not path:
        return None
    return [path]


def find_hifetch_cmd() -> List[str] | None:
    default_path = Path("~/Downloads/hifetch/bin/hifetch").expanduser()
    if default_path.exists() and os.access(default_path, os.X_OK):
        return [str(default_path)]
    return find_cmd("hifetch")


def print_table(results: List[BenchResult], use_color: bool, title: str = "", title_color: str = "") -> None:
    headers = [
        "Tool",
        "Runs",
        "Mean(ms)",
        "Median(ms)",
        "P95(ms)",
        "Min(ms)",
        "Max(ms)",
        "StdDev(ms)",
    ]

    rows = []
    for r in sorted(results, key=lambda x: x.mean_ms):
        rows.append(
            [
                r.name,
                str(r.runs),
                f"{r.mean_ms:.2f}",
                f"{r.median_ms:.2f}",
                f"{r.p95_ms:.2f}",
                f"{r.min_ms:.2f}",
                f"{r.max_ms:.2f}",
                f"{r.stddev_ms:.2f}",
            ]
        )

    widths = [len(h) for h in headers]
    for row in rows:
        for i, col in enumerate(row):
            widths[i] = max(widths[i], len(col))
    table_width = sum(widths) + (2 * (len(widths) - 1))

    def fmt(row: List[str]) -> str:
        return "  ".join(col.ljust(widths[i]) for i, col in enumerate(row))

    if title:
        centered = title.center(table_width)
        color = title_color or C.CYAN
        print(ctext(centered, C.BOLD + color, use_color))
        print()

    print(ctext(fmt(headers), C.BOLD, use_color))
    print(ctext("  ".join("-" * w for w in widths), C.DIM, use_color))
    for i, row in enumerate(rows):
        line = fmt(row)
        if i == 0:
            line = ctext(line, C.GREEN, use_color)
        print(line)


def hyperfine_available() -> bool:
    return shutil.which("hyperfine") is not None


def run_hyperfine_bench(
    repo_root: Path, targets: List[tuple[str, List[str]]], runs: int, warmup: int
) -> tuple[List[BenchResult], str]:
    hf = shutil.which("hyperfine")
    if not hf:
        return [], "hyperfine not found in PATH"

    with tempfile.NamedTemporaryFile(prefix="dawnfetch-hyperfine-", suffix=".json", delete=False) as tmp:
        export_path = Path(tmp.name)

    try:
        cmd = [
            hf,
            "--warmup",
            str(warmup),
            "--runs",
            str(runs),
            "--export-json",
            str(export_path),
        ]
        for name, argv in targets:
            cmd.extend(["-n", name, " ".join(shlex.quote(x) for x in argv)])

        proc = subprocess.run(cmd, cwd=repo_root, capture_output=True, text=True, check=False)
        if proc.returncode != 0:
            msg = (proc.stderr or proc.stdout or "hyperfine failed").strip()
            return [], msg

        data = json.loads(export_path.read_text(encoding="utf-8"))
        raw_results = data.get("results", [])
        out: List[BenchResult] = []
        for i, raw in enumerate(raw_results):
            name = targets[i][0] if i < len(targets) else str(raw.get("command", f"cmd#{i+1}"))
            times_sec = raw.get("times") or []
            samples_ms = [safe_float(v, 0.0) * 1000.0 for v in times_sec if v is not None]
            samples_ms.sort()
            if not samples_ms:
                return [], (
                    f"hyperfine returned no timing samples for {name}. "
                    "try increasing --runs."
                )

            mean_ms = safe_float(raw.get("mean"), statistics.fmean(samples_ms) / 1000.0) * 1000.0
            median_ms = safe_float(raw.get("median"), statistics.median(samples_ms) / 1000.0) * 1000.0
            stddev_ms = safe_float(raw.get("stddev"), 0.0) * 1000.0
            min_ms = safe_float(raw.get("min"), samples_ms[0] / 1000.0) * 1000.0
            max_ms = safe_float(raw.get("max"), samples_ms[-1] / 1000.0) * 1000.0
            p95_ms = percentile(samples_ms, 95) if samples_ms else 0.0
            out.append(
                BenchResult(
                    name=name,
                    runs=len(samples_ms) if samples_ms else runs,
                    mean_ms=mean_ms,
                    median_ms=median_ms,
                    p95_ms=p95_ms,
                    min_ms=min_ms,
                    max_ms=max_ms,
                    stddev_ms=stddev_ms,
                )
            )
        return out, ""
    except Exception as exc:  # pragma: no cover
        return [], str(exc)
    finally:
        try:
            export_path.unlink(missing_ok=True)
        except Exception:
            pass


def safe_float(value: object, fallback: float) -> float:
    if value is None:
        return fallback
    try:
        return float(value)
    except (TypeError, ValueError):
        return fallback


def resolve_targets(repo_root: Path, args: argparse.Namespace) -> List[tuple[str, List[str]]]:
    targets: List[tuple[str, List[str]]] = []

    dawnfetch_cmd = prepare_dawnfetch_cmd(repo_root, args.dawnfetch_cmd)
    targets.append(("dawnfetch", dawnfetch_cmd))

    fastfetch_cmd = find_cmd("fastfetch")
    if fastfetch_cmd:
        targets.append(("fastfetch", fastfetch_cmd + ["--logo", "none"]))
    else:
        print("skip fastfetch: not found in PATH", file=sys.stderr)

    hifetch_cmd = find_hifetch_cmd()
    if hifetch_cmd:
        targets.append(("hifetch", hifetch_cmd))
    else:
        print("skip hifetch: not found at ~/Downloads/hifetch/bin/hifetch or in PATH", file=sys.stderr)

    neofetch_cmd = find_cmd("neofetch")
    if neofetch_cmd:
        targets.append(("neofetch", neofetch_cmd + ["--off"]))
    else:
        print("skip neofetch: not found in PATH", file=sys.stderr)

    screenfetch_cmd = find_cmd("screenfetch")
    if screenfetch_cmd:
        targets.append(("screenfetch", screenfetch_cmd))
    else:
        print("skip screenfetch: not found in PATH", file=sys.stderr)

    macchina_cmd = find_cmd("macchina")
    if macchina_cmd:
        targets.append(("macchina", macchina_cmd))
    else:
        print("skip macchina: not found in PATH", file=sys.stderr)
    return targets


def plot_results(
    results: List[BenchResult], preset: str, runs: int, warmup: int
) -> tuple[Path | None, str]:
    if not results:
        return None, "no benchmark results to plot"
    try:
        import matplotlib

        matplotlib.use("Agg")
        import matplotlib.pyplot as plt
        from matplotlib import ticker as mticker
        import numpy as np
    except Exception as exc:
        return None, f"matplotlib is not available ({exc})"

    ordered = sorted(results, key=lambda x: x.mean_ms)
    tools = [r.name for r in ordered]
    means = np.array([r.mean_ms for r in ordered], dtype=float)

    min_mean = float(means.min())
    max_mean = float(means.max())
    if math.isclose(max_mean, min_mean):
        norm = np.zeros_like(means)
    else:
        norm = (means - min_mean) / (max_mean - min_mean)

    cmap = matplotlib.colormaps.get_cmap("RdYlGn_r")
    colors = [cmap(float(v)) for v in norm]

    # clean and explicit style so output stays consistent across machines.
    plt.style.use("default")
    plt.rcParams.update(
        {
            "figure.facecolor": "#f8fafc",
            "axes.facecolor": "#ffffff",
            "axes.edgecolor": "#dbe4ee",
            "axes.labelcolor": "#0f172a",
            "xtick.color": "#334155",
            "ytick.color": "#0f172a",
            "font.size": 11,
            "grid.color": "#cbd5e1",
            "grid.alpha": 0.35,
            "grid.linestyle": "--",
        }
    )

    fig_h = max(5.2, 0.72 * len(tools) + 2.8)
    fig_w = 14.2
    fig, ax = plt.subplots(figsize=(fig_w, fig_h))
    y = np.arange(len(tools))

    bars = ax.barh(
        y,
        means,
        color=colors,
        edgecolor="#0f172a",
        linewidth=0.7,
        height=0.72,
        zorder=2,
    )
    ax.invert_yaxis()

    x_peak = float(np.max(means))
    x_pad = max(30.0, x_peak * 0.20)
    right_limit = x_peak + x_pad
    ax.set_xlim(0, right_limit)
    ax.set_yticks(y, labels=tools)
    ax.tick_params(axis="y", pad=6)
    for label in ax.get_yticklabels():
        label.set_clip_on(False)
    ax.set_xlabel("runtime (ms)", fontsize=12, fontweight="normal")
    ax.set_title(
        f"dawnfetch benchmark results ({preset} | runs={runs}, warmup={warmup})",
        pad=15,
        fontsize=17,
        fontweight="bold",
        color="#0f172a",
    )
    ax.grid(axis="x", zorder=1)
    ax.set_axisbelow(True)
    ax.xaxis.set_major_locator(mticker.MaxNLocator(nbins=10))
    ax.xaxis.set_major_formatter(mticker.StrMethodFormatter("{x:.0f}"))
    # subtle frame lines for cleaner boundaries.
    for spine in ax.spines.values():
        spine.set_linewidth(1.2)
        spine.set_edgecolor("#0f172a")

    # annotate mean values only (execution-time chart).
    label_x_gap = x_pad * 0.05
    for bar, m in zip(bars, means):
        y_pos = bar.get_y() + bar.get_height() / 2
        x_pos = min(m + label_x_gap, right_limit - (x_pad * 0.02))
        ax.text(
            x_pos,
            y_pos,
            f"{m:.2f} ms",
            va="center",
            ha="left",
            fontsize=10,
            color="#0f172a",
            clip_on=True,
            zorder=5,
        )

    fastest = ordered[0]
    slowest = ordered[-1]
    summary = (
        f"fastest: {fastest.name} ({fastest.mean_ms:.2f}ms)  |  "
        f"slowest: {slowest.name} ({slowest.mean_ms:.2f}ms)"
    )
    fig.text(0.955, 0.03, summary, fontsize=10, color="#334155", fontweight="bold", ha="right")

    plots_dir = Path(__file__).resolve().parent / "plots"
    try:
        plots_dir.mkdir(parents=True, exist_ok=True)
    except OSError as exc:
        return None, f"failed to create plots directory at {plots_dir}: {exc}"
    out_path = plots_dir / f"benchmark_plot_{dt.datetime.now().strftime('%Y%m%d_%H%M%S')}.png"
    # fixed margins keep outer padding visually balanced and predictable.
    fig.subplots_adjust(left=0.10, right=0.965, top=0.90, bottom=0.12)
    fig.savefig(out_path, dpi=190, facecolor=fig.get_facecolor())
    plt.close(fig)
    return out_path, ""


def main() -> int:
    args = parse_args()
    runs, warmup = runs_warmup_from_args(args)
    if runs <= 0 or warmup < 0:
        print("runs must be > 0 and warmup must be >= 0", file=sys.stderr)
        return 2
    if not args.allow_small_sample and (runs < 10 or warmup < 1):
        print(
            "refusing unreliable benchmark config: use runs >= 10 and warmup >= 1 "
            "(or pass --allow-small-sample to override)",
            file=sys.stderr,
        )
        return 2
    use_color = color_enabled(args.force_color, args.no_color)

    repo_root = Path(__file__).resolve().parent.parent
    try:
        targets = resolve_targets(repo_root, args)
    except RuntimeError as exc:
        print(str(exc), file=sys.stderr)
        return 1

    if len(targets) == 0:
        print("No benchmark targets available.", file=sys.stderr)
        return 1

    print(ctext(f"preset: {args.preset} | runs={runs}, warmup={warmup}", C.DIM, use_color))
    if not hyperfine_available():
        print("hyperfine requested but not found in PATH", file=sys.stderr)
        return 1

    print()
    for name, cmd in targets:
        print(
            f"{ctext('>>', C.BLUE, use_color)} running {ctext(name, C.BOLD, use_color)}: "
            f"{' '.join(shlex.quote(c) for c in cmd)}"
        )

    results, hf_err = run_hyperfine_bench(repo_root, targets, runs, warmup)
    print()
    if not results:
        msg = hf_err or "unable to collect hyperfine results"
        print(ctext(f"error: {msg}", C.RED, use_color), file=sys.stderr)
        return 1
    print_table(results, use_color, "hyperfine benchmark", C.CYAN)

    if args.plot == "on":
        out_path, plot_err = plot_results(results, args.preset, runs, warmup)
        if out_path is not None:
            print()
            print(ctext(f"plot saved: {clickable_path(out_path)}", C.DIM, use_color))
        else:
            print(ctext(f"plot skipped: {plot_err}", C.RED, use_color), file=sys.stderr)
            return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
