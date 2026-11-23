// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

interface ReadFile {
  name: string;
  mode: "embed" | "runtime";
}

interface CellInfo {
  name: string;
  line: number;
  exported: string;
  body: string;
  reads: ReadFile[];
  writes: string[];
  secrets: string[];
  data: string;
  timeout: string;
}

interface UpdateRequest {
  line: number;
  body: string;
  name: string;
  exported: string;
  reads: ReadFile[];
  writes: string[];
  secrets: string[];
  data: string;
  timeout: string;
}

interface ApiError {
  error: string;
}

class CellEditor {
  private container: HTMLElement;

  constructor(container: HTMLElement) {
    this.container = container;
  }

  async load(): Promise<void> {
    const resp = await fetch("/api/cells");
    if (!resp.ok) {
      const err: ApiError = await resp.json() as ApiError;
      this.showError("Failed to load cells: " + err.error);
      return;
    }
    const cells: CellInfo[] = await resp.json() as CellInfo[];
    this.render(cells);
  }

  private render(cells: CellInfo[]): void {
    this.container.innerHTML = "";
    if (cells.length === 0) {
      const p = document.createElement("p");
      p.textContent = "No cells found in notebook.";
      this.container.appendChild(p);
      return;
    }
    for (const cell of cells) {
      this.container.appendChild(this.buildCellForm(cell));
    }
  }

  private buildCellForm(cell: CellInfo): HTMLElement {
    const section = document.createElement("section");
    section.className = "cell";
    section.dataset["line"] = String(cell.line);

    const heading = document.createElement("h2");
    heading.textContent = cell.name ? "Cell: " + cell.name : "Cell (line " + cell.line + ")";
    section.appendChild(heading);

    // Status message area
    const status = document.createElement("p");
    status.className = "status";
    section.appendChild(status);

    // body
    section.appendChild(this.labeledTextarea("Body", "body", cell.body, true));
    // name
    section.appendChild(this.labeledInput("Name", "name", cell.name));
    // exported
    section.appendChild(this.labeledInput("Exported", "exported", cell.exported));
    // reads
    section.appendChild(this.labeledInput("Reads (comma-separated, name:mode)", "reads", this.readsToString(cell.reads)));
    // writes
    section.appendChild(this.labeledInput("Writes (comma-separated)", "writes", (cell.writes ?? []).join(", ")));
    // secrets
    section.appendChild(this.labeledInput("Secrets (comma-separated)", "secrets", (cell.secrets ?? []).join(", ")));
    // data
    const dataField = this.labeledSelect("Data", "data", ["", "embedded", "live"], cell.data ?? "");
    section.appendChild(dataField);
    // timeout
    section.appendChild(this.labeledInput("Timeout (e.g. 30s, 2m)", "timeout", cell.timeout ?? ""));

    // Save button
    const saveBtn = document.createElement("button");
    saveBtn.type = "button";
    saveBtn.textContent = "Save";
    saveBtn.addEventListener("click", () => {
      void this.saveCell(section, cell.line, status, heading);
    });
    section.appendChild(saveBtn);

    return section;
  }

  private labeledInput(label: string, name: string, value: string): HTMLElement {
    const div = document.createElement("div");
    div.className = "field";
    const lbl = document.createElement("label");
    lbl.textContent = label;
    const input = document.createElement("input");
    input.type = "text";
    input.name = name;
    input.value = value ?? "";
    div.appendChild(lbl);
    div.appendChild(input);
    return div;
  }

  private labeledTextarea(label: string, name: string, value: string, monospace: boolean): HTMLElement {
    const div = document.createElement("div");
    div.className = "field";
    const lbl = document.createElement("label");
    lbl.textContent = label;
    const ta = document.createElement("textarea");
    ta.name = name;
    ta.value = value ?? "";
    ta.rows = 8;
    if (monospace) {
      ta.className = "mono";
    }
    div.appendChild(lbl);
    div.appendChild(ta);
    return div;
  }

  private labeledSelect(label: string, name: string, options: string[], selected: string): HTMLElement {
    const div = document.createElement("div");
    div.className = "field";
    const lbl = document.createElement("label");
    lbl.textContent = label;
    const sel = document.createElement("select");
    sel.name = name;
    for (const opt of options) {
      const o = document.createElement("option");
      o.value = opt;
      o.textContent = opt === "" ? "(none)" : opt;
      if (opt === (selected ?? "")) {
        o.selected = true;
      }
      sel.appendChild(o);
    }
    div.appendChild(lbl);
    div.appendChild(sel);
    return div;
  }

  private readsToString(reads: ReadFile[]): string {
    if (!reads || reads.length === 0) return "";
    return reads.map(r => r.name + ":" + r.mode).join(", ");
  }

  private parseReads(raw: string): ReadFile[] {
    if (!raw.trim()) return [];
    return raw.split(",").map(s => s.trim()).filter(s => s.length > 0).map(s => {
      const idx = s.lastIndexOf(":");
      if (idx < 0) return { name: s, mode: "embed" as const };
      const mode = s.slice(idx + 1).trim();
      return {
        name: s.slice(0, idx).trim(),
        mode: (mode === "runtime" ? "runtime" : "embed") as "embed" | "runtime",
      };
    });
  }

  private parseCSV(raw: string): string[] {
    if (!raw.trim()) return [];
    return raw.split(",").map(s => s.trim()).filter(s => s.length > 0);
  }

  private getValue(section: HTMLElement, name: string): string {
    const el = section.querySelector<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>(`[name="${name}"]`);
    return el ? el.value : "";
  }

  private async saveCell(
    section: HTMLElement,
    line: number,
    status: HTMLElement,
    heading: HTMLElement,
  ): Promise<void> {
    status.textContent = "Saving…";
    status.className = "status";

    const req: UpdateRequest = {
      line,
      body: this.getValue(section, "body"),
      name: this.getValue(section, "name"),
      exported: this.getValue(section, "exported"),
      reads: this.parseReads(this.getValue(section, "reads")),
      writes: this.parseCSV(this.getValue(section, "writes")),
      secrets: this.parseCSV(this.getValue(section, "secrets")),
      data: this.getValue(section, "data"),
      timeout: this.getValue(section, "timeout"),
    };

    try {
      const resp = await fetch("/api/cells/update", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(req),
      });

      if (!resp.ok) {
        const err: ApiError = await resp.json() as ApiError;
        status.textContent = "Error: " + err.error;
        status.className = "status error";
        return;
      }

      status.textContent = "Saved.";
      status.className = "status ok";
      // Update heading to reflect possible name change.
      const newName = req.name;
      heading.textContent = newName ? "Cell: " + newName : "Cell (line " + line + ")";
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      status.textContent = "Network error: " + msg;
      status.className = "status error";
    }
  }

  private showError(msg: string): void {
    const p = document.createElement("p");
    p.className = "status error";
    p.textContent = msg;
    this.container.appendChild(p);
  }
}

document.addEventListener("DOMContentLoaded", () => {
  const container = document.getElementById("cells");
  if (!container) return;
  const editor = new CellEditor(container);
  void editor.load();
});
