// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Chart interactivity script for go-notebook HTML reports.
// Reads chart data from data-chart attributes on SVG elements and adds:
// - Hover tooltips showing values
// - Zoom (double-click to reset)
// - Series toggling via legend clicks
//
// This script never evaluates data as code.

interface ChartSeries {
  name: string;
  values: number[];
}

interface ChartData {
  type: string;
  title: string;
  labels: string[];
  x: number[];
  series: ChartSeries[];
}

interface TooltipState {
  el: HTMLDivElement;
  visible: boolean;
}

class ChartInteraction {
  private svg: SVGSVGElement;
  private data: ChartData;
  private tooltip: TooltipState;
  private hiddenSeries: Set<string>;
  private zoomLevel: number;
  private zoomOriginX: number;
  private zoomOriginY: number;

  constructor(svg: SVGSVGElement, data: ChartData) {
    this.svg = svg;
    this.data = data;
    this.hiddenSeries = new Set<string>();
    this.zoomLevel = 1;
    this.zoomOriginX = 0;
    this.zoomOriginY = 0;

    // Create tooltip element.
    const tip = document.createElement("div");
    tip.className = "chart-tooltip";
    tip.style.cssText = "position:absolute;display:none;background:#333;color:#fff;padding:4px 8px;border-radius:4px;font-size:12px;pointer-events:none;z-index:1000;white-space:nowrap;";
    document.body.appendChild(tip);
    this.tooltip = { el: tip, visible: false };

    this.setupHover();
    this.setupZoom();
    this.setupLegendToggle();
  }

  private setupHover(): void {
    // Add hover listeners to data elements (rects, circles, paths).
    const dataElements = this.svg.querySelectorAll("rect, circle, path[d]");
    for (const el of dataElements) {
      el.addEventListener("mouseenter", (e: Event) => {
        const target = e.target as SVGElement;
        this.showTooltip(target, e as MouseEvent);
      });
      el.addEventListener("mousemove", (e: Event) => {
        this.moveTooltip(e as MouseEvent);
      });
      el.addEventListener("mouseleave", () => {
        this.hideTooltip();
      });
    }
  }

  private showTooltip(el: SVGElement, e: MouseEvent): void {
    // Extract value from element attributes or data.
    let text = "";
    const title = el.querySelector("title");
    if (title) {
      text = title.textContent ?? "";
    } else {
      // Try to infer from element position and chart data.
      const fill = el.getAttribute("fill") ?? "";
      const seriesIdx = this.getSeriesIndexByColor(fill);
      if (seriesIdx >= 0 && seriesIdx < this.data.series.length) {
        text = this.data.series[seriesIdx].name;
      }
    }

    if (text) {
      this.tooltip.el.textContent = text;
      this.tooltip.el.style.display = "block";
      this.tooltip.visible = true;
      this.moveTooltip(e);
    }
  }

  private moveTooltip(e: MouseEvent): void {
    if (!this.tooltip.visible) return;
    this.tooltip.el.style.left = (e.pageX + 12) + "px";
    this.tooltip.el.style.top = (e.pageY - 20) + "px";
  }

  private hideTooltip(): void {
    this.tooltip.el.style.display = "none";
    this.tooltip.visible = false;
  }

  private setupZoom(): void {
    // Click to zoom in, double-click to reset.
    this.svg.addEventListener("click", (e: MouseEvent) => {
      if (this.zoomLevel >= 4) return; // Max zoom.
      const rect = this.svg.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;

      this.zoomLevel *= 2;
      this.zoomOriginX = x;
      this.zoomOriginY = y;
      this.applyZoom();
    });

    this.svg.addEventListener("dblclick", () => {
      this.zoomLevel = 1;
      this.zoomOriginX = 0;
      this.zoomOriginY = 0;
      this.applyZoom();
    });
  }

  private applyZoom(): void {
    if (this.zoomLevel === 1) {
      this.svg.style.transform = "";
      this.svg.style.transformOrigin = "";
      return;
    }
    this.svg.style.transformOrigin = this.zoomOriginX + "px " + this.zoomOriginY + "px";
    this.svg.style.transform = "scale(" + this.zoomLevel + ")";
  }

  private setupLegendToggle(): void {
    // Find legend text elements (they match series names).
    const textElements = this.svg.querySelectorAll("text");
    for (const textEl of textElements) {
      const name = textEl.textContent?.trim() ?? "";
      const matchingSeries = this.data.series.find(s => s.name === name);
      if (matchingSeries) {
        textEl.style.cursor = "pointer";
        textEl.addEventListener("click", (e: Event) => {
          e.stopPropagation(); // Don't trigger zoom.
          this.toggleSeries(name);
        });
      }
    }
  }

  private toggleSeries(name: string): void {
    if (this.hiddenSeries.has(name)) {
      this.hiddenSeries.delete(name);
    } else {
      this.hiddenSeries.add(name);
    }

    // Find the series index and color, then hide/show its elements.
    const idx = this.data.series.findIndex(s => s.name === name);
    if (idx < 0) return;

    const colors = ["#4e79a7", "#f28e2b", "#e15759", "#76b7b2", "#59a14f", "#edc948"];
    const color = colors[idx % colors.length];
    const hidden = this.hiddenSeries.has(name);

    // Find elements with this color and toggle visibility.
    const elements = this.svg.querySelectorAll(`[fill="${color}"], [stroke="${color}"]`);
    for (const el of elements) {
      (el as SVGElement).style.opacity = hidden ? "0.1" : "1";
    }

    // Update legend text style.
    const textElements = this.svg.querySelectorAll("text");
    for (const textEl of textElements) {
      if (textEl.textContent?.trim() === name) {
        textEl.style.textDecoration = hidden ? "line-through" : "none";
        textEl.style.opacity = hidden ? "0.5" : "1";
      }
    }
  }

  private getSeriesIndexByColor(fill: string): number {
    const colors = ["#4e79a7", "#f28e2b", "#e15759", "#76b7b2", "#59a14f", "#edc948"];
    return colors.indexOf(fill);
  }
}

// Initialize all charts on the page.
function initCharts(): void {
  const svgs = document.querySelectorAll<SVGSVGElement>("svg[data-chart]");
  for (const svg of svgs) {
    const raw = svg.getAttribute("data-chart");
    if (!raw) continue;
    try {
      const data: ChartData = JSON.parse(atob(raw)) as ChartData;
      new ChartInteraction(svg, data);
    } catch {
      // Invalid chart data — skip silently.
    }
  }
}

// Run on DOM ready.
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initCharts);
} else {
  initCharts();
}
