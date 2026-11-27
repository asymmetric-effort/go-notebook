// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

interface ReadFile {
  name: string;
  mode: "embed" | "runtime";
}

interface BlockInfo {
  kind: string;
  line: number;
  end_line: number;
  body: string;
  info: string;
  name: string;
  exported: string;
}

interface CellUpdateRequest {
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

interface ProseUpdateRequest {
  line: number;
  body: string;
}

interface ApiError {
  error: string;
}

class NotebookEditor {
  private container: HTMLElement;

  constructor(container: HTMLElement) {
    this.container = container;
  }

  async load(): Promise<void> {
    const resp = await fetch("/api/blocks");
    if (!resp.ok) {
      const err: ApiError = await resp.json() as ApiError;
      this.showError("Failed to load notebook: " + err.error);
      return;
    }
    const blocks: BlockInfo[] = await resp.json() as BlockInfo[];
    this.render(blocks);
  }

  private render(blocks: BlockInfo[]): void {
    this.container.innerHTML = "";
    if (!blocks || blocks.length === 0) {
      this.container.appendChild(this.buildGetStarted());
      return;
    }

    // Run All button at the top.
    const toolbar = document.createElement("div");
    toolbar.className = "toolbar";
    const runAllBtn = document.createElement("button");
    runAllBtn.type = "button";
    runAllBtn.textContent = "Run All";
    runAllBtn.className = "run-all-btn";
    const runAllStatus = document.createElement("span");
    runAllStatus.className = "status";
    runAllBtn.addEventListener("click", () => {
      void this.runAll(runAllStatus);
    });
    toolbar.appendChild(runAllBtn);
    toolbar.appendChild(runAllStatus);
    this.container.appendChild(toolbar);

    // Secret prompt section.
    this.container.appendChild(this.buildSecretPrompt());

    // Insert toolbar at the top (before first block).
    this.container.appendChild(this.buildInsertToolbar(0));

    for (let i = 0; i < blocks.length; i++) {
      const block = blocks[i];
      const isFirst = i === 0;
      const isLast = i === blocks.length - 1;
      switch (block.kind) {
        case "prose":
          this.container.appendChild(this.buildProseForm(block, isFirst, isLast));
          break;
        case "cell":
          this.container.appendChild(this.buildCellForm(block, isFirst, isLast));
          break;
        case "defs":
          this.container.appendChild(this.buildDefsBlock(block, isFirst, isLast));
          break;
        case "output":
          this.container.appendChild(this.buildOutputBlock(block, isFirst, isLast));
          break;
      }
      // Insert toolbar after each block.
      this.container.appendChild(this.buildInsertToolbar(block.end_line));
    }

    // Load initial cell states and diagnostics.
    void this.refreshStaleMarkers();
    void this.refreshDiagnostics();

    // Set up keyboard shortcuts.
    this.setupKeyboardShortcuts();
  }

  private buildGetStarted(): HTMLElement {
    const div = document.createElement("div");
    div.className = "get-started";

    const h2 = document.createElement("h2");
    h2.textContent = "Empty Notebook";
    div.appendChild(h2);

    const p = document.createElement("p");
    p.textContent = "This notebook is empty. Add content to get started.";
    div.appendChild(p);

    const starterBtn = document.createElement("button");
    starterBtn.type = "button";
    starterBtn.textContent = "Create Starter Notebook";
    starterBtn.className = "starter-btn";
    starterBtn.addEventListener("click", () => {
      void this.initNotebook();
    });
    div.appendChild(starterBtn);

    const addRow = document.createElement("div");
    addRow.className = "insert-toolbar";

    const addProse = document.createElement("button");
    addProse.type = "button";
    addProse.textContent = "+ Add Prose";
    addProse.addEventListener("click", () => {
      void this.addBlock("prose", 0);
    });
    addRow.appendChild(addProse);

    const addCell = document.createElement("button");
    addCell.type = "button";
    addCell.textContent = "+ Add Cell";
    addCell.addEventListener("click", () => {
      void this.addBlock("cell", 0);
    });
    addRow.appendChild(addCell);

    const addDefs = document.createElement("button");
    addDefs.type = "button";
    addDefs.textContent = "+ Add Defs";
    addDefs.addEventListener("click", () => {
      void this.addBlock("defs", 0);
    });
    addRow.appendChild(addDefs);

    div.appendChild(addRow);
    return div;
  }

  private buildInsertToolbar(afterLine: number): HTMLElement {
    const div = document.createElement("div");
    div.className = "insert-toolbar";

    const proseBtn = document.createElement("button");
    proseBtn.type = "button";
    proseBtn.textContent = "+ Prose";
    proseBtn.addEventListener("click", () => {
      void this.addBlock("prose", afterLine);
    });
    div.appendChild(proseBtn);

    const cellBtn = document.createElement("button");
    cellBtn.type = "button";
    cellBtn.textContent = "+ Cell";
    cellBtn.addEventListener("click", () => {
      void this.addBlock("cell", afterLine);
    });
    div.appendChild(cellBtn);

    const defsBtn = document.createElement("button");
    defsBtn.type = "button";
    defsBtn.textContent = "+ Defs";
    defsBtn.addEventListener("click", () => {
      void this.addBlock("defs", afterLine);
    });
    div.appendChild(defsBtn);

    return div;
  }

  private buildBlockActions(block: BlockInfo, isFirst: boolean, isLast: boolean): HTMLElement {
    const div = document.createElement("div");
    div.className = "block-actions";

    if (!isFirst) {
      const upBtn = document.createElement("button");
      upBtn.type = "button";
      upBtn.textContent = "\u2191";
      upBtn.title = "Move up";
      upBtn.addEventListener("click", () => {
        void this.reorderBlock(block.line, "up");
      });
      div.appendChild(upBtn);
    }

    if (!isLast) {
      const downBtn = document.createElement("button");
      downBtn.type = "button";
      downBtn.textContent = "\u2193";
      downBtn.title = "Move down";
      downBtn.addEventListener("click", () => {
        void this.reorderBlock(block.line, "down");
      });
      div.appendChild(downBtn);
    }

    const deleteBtn = document.createElement("button");
    deleteBtn.type = "button";
    deleteBtn.textContent = "Delete";
    deleteBtn.className = "delete-btn";
    deleteBtn.addEventListener("click", () => {
      if (confirm("Delete this block?")) {
        void this.deleteBlock(block.line);
      }
    });
    div.appendChild(deleteBtn);

    return div;
  }

  private async addBlock(kind: string, afterLine: number): Promise<void> {
    try {
      const body: Record<string, unknown> = { kind, after_line: afterLine, body: "" };
      if (kind === "cell") {
        body.info = "go:notebook name=cell1";
      } else if (kind === "defs") {
        body.info = "go:defs";
      }
      const resp = await fetch("/api/blocks/add", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!resp.ok) {
        const err: ApiError = await resp.json() as ApiError;
        alert("Error adding block: " + err.error);
        return;
      }
      void this.load();
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      alert("Network error: " + msg);
    }
  }

  private async deleteBlock(line: number): Promise<void> {
    try {
      const resp = await fetch("/api/blocks/delete", {
        method: "DELETE",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ line }),
      });
      if (!resp.ok) {
        const err: ApiError = await resp.json() as ApiError;
        alert("Error deleting block: " + err.error);
        return;
      }
      void this.load();
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      alert("Network error: " + msg);
    }
  }

  private async reorderBlock(line: number, direction: string): Promise<void> {
    try {
      const resp = await fetch("/api/blocks/reorder", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ line, direction }),
      });
      if (!resp.ok) {
        const err: ApiError = await resp.json() as ApiError;
        alert("Error reordering block: " + err.error);
        return;
      }
      void this.load();
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      alert("Network error: " + msg);
    }
  }

  private async initNotebook(): Promise<void> {
    try {
      const resp = await fetch("/api/notebook/init", {
        method: "POST",
      });
      if (!resp.ok) {
        const err: ApiError = await resp.json() as ApiError;
        alert("Error: " + err.error);
        return;
      }
      void this.load();
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      alert("Network error: " + msg);
    }
  }

  private setupKeyboardShortcuts(): void {
    const textareas = this.container.querySelectorAll<HTMLTextAreaElement>("textarea");
    for (const ta of textareas) {
      ta.addEventListener("keydown", (e: KeyboardEvent) => {
        // Shift+Enter: run cell
        if (e.key === "Enter" && e.shiftKey && !e.ctrlKey && !e.altKey && !e.metaKey) {
          e.preventDefault();
          const section = ta.closest("section");
          if (section) {
            const runBtn = section.querySelector<HTMLButtonElement>(".run-btn");
            if (runBtn) {
              runBtn.click();
            }
          }
        }
        // Ctrl+S / Cmd+S: save
        if (e.key === "s" && (e.ctrlKey || e.metaKey) && !e.shiftKey && !e.altKey) {
          e.preventDefault();
          const section = ta.closest("section");
          if (section) {
            // Find the save button — it's the first button whose text contains "Save".
            const buttons = section.querySelectorAll<HTMLButtonElement>("button");
            for (const btn of buttons) {
              if (btn.textContent && btn.textContent.includes("Save")) {
                btn.click();
                break;
              }
            }
          }
        }
      });
    }
  }

  private async refreshDiagnostics(): Promise<void> {
    try {
      const resp = await fetch("/api/lsp/diagnostics");
      if (!resp.ok) return;
      interface Diagnostic {
        file: string;
        line: number;
        col: number;
        message: string;
        severity: string;
      }
      const data: { diagnostics: Diagnostic[] } = await resp.json() as { diagnostics: Diagnostic[] };
      const diags = data.diagnostics ?? [];

      // Clear existing diagnostic markers.
      const existing = this.container.querySelectorAll(".diagnostic-marker");
      for (const el of existing) {
        el.remove();
      }

      // Show diagnostics panel if there are any.
      let panel = document.getElementById("diagnostics-panel");
      if (!panel) {
        panel = document.createElement("div");
        panel.id = "diagnostics-panel";
        panel.className = "diagnostics-panel";
        this.container.insertBefore(panel, this.container.firstChild);
      }

      if (diags.length === 0) {
        panel.style.display = "none";
        return;
      }

      panel.style.display = "block";
      panel.textContent = "";
      const heading = document.createElement("h3");
      heading.textContent = "Diagnostics (" + diags.length + ")";
      panel.appendChild(heading);

      for (const d of diags) {
        const div = document.createElement("div");
        div.className = "diagnostic diagnostic-" + d.severity;
        div.textContent = (d.line > 0 ? "Line " + d.line + ": " : "") + d.message;
        panel.appendChild(div);
      }
    } catch {
      // Ignore diagnostic fetch errors.
    }
  }

  private async runAll(status: HTMLElement): Promise<void> {
    status.textContent = " Running all cells…";
    status.className = "status";
    try {
      const resp = await fetch("/api/run/all", { method: "POST" });
      if (!resp.ok) {
        const err: ApiError = await resp.json() as ApiError;
        status.textContent = " Error: " + err.error;
        status.className = "status error";
        return;
      }
      status.textContent = " Done.";
      status.className = "status ok";
      void this.refreshStaleMarkers();
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      status.textContent = " Network error: " + msg;
      status.className = "status error";
    }
  }

  private async refreshStaleMarkers(): Promise<void> {
    try {
      const resp = await fetch("/api/run/state");
      if (!resp.ok) return;
      const states: Record<string, { stale: boolean; failed: boolean; stdout: string }> =
        await resp.json() as Record<string, { stale: boolean; failed: boolean; stdout: string }>;
      // Update stale/failed markers on cell sections.
      const sections = this.container.querySelectorAll<HTMLElement>("section.cell-block");
      for (const sec of sections) {
        const name = sec.dataset["cellName"] ?? "";
        const st = states[name];
        sec.classList.remove("stale", "failed");
        const marker = sec.querySelector(".cell-marker");
        if (st) {
          if (st.stale) {
            sec.classList.add("stale");
            if (marker) marker.textContent = " (stale)";
          } else if (st.failed) {
            sec.classList.add("failed");
            if (marker) marker.textContent = " (failed)";
          } else {
            if (marker) marker.textContent = " (ok)";
          }
        } else {
          if (marker) marker.textContent = "";
        }
      }
    } catch {
      // Ignore state fetch errors.
    }
  }

  // --- Secret prompt ---

  private buildSecretPrompt(): HTMLElement {
    const section = document.createElement("section");
    section.className = "block secret-block";
    section.style.display = "none"; // Hidden by default, shown when needed.
    section.id = "secret-prompt";

    const heading = document.createElement("h3");
    heading.textContent = "Secret";
    section.appendChild(heading);

    const status = document.createElement("p");
    status.className = "status";
    section.appendChild(status);

    const nameField = document.createElement("div");
    nameField.className = "field";
    const nameLabel = document.createElement("label");
    nameLabel.textContent = "Secret Name";
    const nameInput = document.createElement("input");
    nameInput.type = "text";
    nameInput.name = "secret-name";
    nameInput.placeholder = "e.g., api_key";
    nameField.appendChild(nameLabel);
    nameField.appendChild(nameInput);
    section.appendChild(nameField);

    const valueField = document.createElement("div");
    valueField.className = "field";
    const valueLabel = document.createElement("label");
    valueLabel.textContent = "Secret Value (masked)";
    const valueInput = document.createElement("input");
    valueInput.type = "password";
    valueInput.name = "secret-value";
    valueInput.autocomplete = "off";
    valueField.appendChild(valueLabel);
    valueField.appendChild(valueInput);
    section.appendChild(valueField);

    const submitBtn = document.createElement("button");
    submitBtn.type = "button";
    submitBtn.textContent = "Set Secret";
    submitBtn.addEventListener("click", () => {
      void this.submitSecret(nameInput, valueInput, status);
    });
    section.appendChild(submitBtn);

    // Always visible toggle.
    section.style.display = "block";

    return section;
  }

  private async submitSecret(
    nameInput: HTMLInputElement,
    valueInput: HTMLInputElement,
    status: HTMLElement,
  ): Promise<void> {
    const name = nameInput.value.trim();
    const value = valueInput.value;

    if (!name) {
      status.textContent = "Secret name is required.";
      status.className = "status error";
      return;
    }

    status.textContent = "Setting…";
    status.className = "status";

    try {
      const resp = await fetch("/api/secrets", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, value }),
      });

      if (!resp.ok) {
        interface ErrResp { error: string }
        const err: ErrResp = await resp.json() as ErrResp;
        status.textContent = "Error: " + err.error;
        status.className = "status error";
        return;
      }

      status.textContent = "Secret \"" + name + "\" set (memory only).";
      status.className = "status ok";
      // Clear the value input for security.
      valueInput.value = "";
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      status.textContent = "Network error: " + msg;
      status.className = "status error";
    }
  }

  // --- Prose editing ---

  private buildProseForm(block: BlockInfo, isFirst: boolean = false, isLast: boolean = false): HTMLElement {
    const section = document.createElement("section");
    section.className = "block prose-block";

    const heading = document.createElement("h3");
    heading.appendChild(this.buildBlockActions(block, isFirst, isLast));
    heading.appendChild(document.createTextNode("Prose (line " + block.line + ")"));
    section.appendChild(heading);

    const status = document.createElement("p");
    status.className = "status";
    section.appendChild(status);

    const ta = document.createElement("textarea");
    ta.name = "prose-body";
    ta.value = block.body;
    ta.rows = Math.max(3, block.body.split("\n").length + 1);
    ta.className = "prose-editor";
    section.appendChild(ta);

    const saveBtn = document.createElement("button");
    saveBtn.type = "button";
    saveBtn.textContent = "Save Prose";
    saveBtn.addEventListener("click", () => {
      void this.saveProse(block.line, ta, status);
    });
    section.appendChild(saveBtn);

    return section;
  }

  private async saveProse(line: number, ta: HTMLTextAreaElement, status: HTMLElement): Promise<void> {
    status.textContent = "Saving…";
    status.className = "status";

    const req: ProseUpdateRequest = {
      line,
      body: ta.value,
    };

    try {
      const resp = await fetch("/api/prose/update", {
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
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      status.textContent = "Network error: " + msg;
      status.className = "status error";
    }
  }

  // --- Cell editing ---

  private buildCellForm(block: BlockInfo, isFirst: boolean = false, isLast: boolean = false): HTMLElement {
    const section = document.createElement("section");
    section.className = "block cell-block";
    section.dataset["line"] = String(block.line);
    section.dataset["cellName"] = block.name ?? "";

    const heading = document.createElement("h2");
    heading.appendChild(this.buildBlockActions(block, isFirst, isLast));
    heading.appendChild(document.createTextNode(block.name ? "Cell: " + block.name : "Cell (line " + block.line + ")"));
    const marker = document.createElement("span");
    marker.className = "cell-marker";
    heading.appendChild(marker);
    section.appendChild(heading);

    const status = document.createElement("p");
    status.className = "status";
    section.appendChild(status);

    // Output area for cell run results.
    const outputArea = document.createElement("pre");
    outputArea.className = "cell-output";
    outputArea.style.display = "none";

    // Display area for rich output (markdown, images, charts).
    const displayArea = document.createElement("div");
    displayArea.className = "cell-display";
    displayArea.style.display = "none";

    // body
    section.appendChild(this.labeledTextarea("Body", "body", block.body, true));
    // name
    section.appendChild(this.labeledInput("Name", "name", block.name));
    // exported
    section.appendChild(this.labeledInput("Exported", "exported", block.exported));
    // reads
    section.appendChild(this.labeledInput("Reads (comma-separated, name:mode)", "reads", ""));
    // writes
    section.appendChild(this.labeledInput("Writes (comma-separated)", "writes", ""));
    // secrets
    section.appendChild(this.labeledInput("Secrets (comma-separated)", "secrets", ""));
    // data
    section.appendChild(this.labeledSelect("Data", "data", ["", "embedded", "live"], ""));
    // timeout
    section.appendChild(this.labeledInput("Timeout (e.g. 30s, 2m)", "timeout", ""));

    // Button row.
    const btnRow = document.createElement("div");
    btnRow.className = "btn-row";

    const saveBtn = document.createElement("button");
    saveBtn.type = "button";
    saveBtn.textContent = "Save Cell";
    saveBtn.addEventListener("click", () => {
      void this.saveCell(section, block.line, status, heading);
    });
    btnRow.appendChild(saveBtn);

    if (block.name) {
      const runBtn = document.createElement("button");
      runBtn.type = "button";
      runBtn.textContent = "Run Cell";
      runBtn.className = "run-btn";
      runBtn.addEventListener("click", () => {
        void this.runCell(block.name, status, outputArea, displayArea);
      });
      btnRow.appendChild(runBtn);
    }

    section.appendChild(btnRow);
    section.appendChild(displayArea);
    section.appendChild(outputArea);

    return section;
  }

  private async runCell(name: string, status: HTMLElement, outputArea: HTMLPreElement, displayArea: HTMLDivElement): Promise<void> {
    status.textContent = "Running…";
    status.className = "status";
    displayArea.style.display = "none";
    displayArea.textContent = "";
    try {
      const resp = await fetch("/api/run", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ cell: name }),
      });
      interface DisplayItem {
        mime: string;
        data: string;
      }
      interface RunResult {
        stdout: string;
        stderr: string;
        exit_code: number;
        error: string;
        refused: string;
        errors: Array<{ message: string; cell: string; file: string; line: number; kind: string }>;
        displays: DisplayItem[];
      }
      const result: RunResult = await resp.json() as RunResult;

      if (result.refused) {
        status.textContent = "Refused: upstream \"" + result.refused + "\" has failed.";
        status.className = "status error";
        return;
      }

      // Display parsed errors if present.
      const errors = result.errors ?? [];
      if (errors.length > 0) {
        this.displayErrors(outputArea, errors);
        status.textContent = errors.length + " error(s)";
        status.className = "status error";
      } else if (result.error) {
        status.textContent = "Error: " + result.error;
        status.className = "status error";
        outputArea.textContent = result.error;
        outputArea.style.display = "block";
      } else if (result.exit_code !== 0) {
        status.textContent = "Exit code: " + result.exit_code;
        status.className = "status error";
        outputArea.textContent = result.stderr || result.stdout;
        outputArea.style.display = "block";
      } else {
        status.textContent = "Done.";
        status.className = "status ok";
        // Show stdout output.
        const output = result.stdout ?? "";
        if (output) {
          outputArea.textContent = output;
          outputArea.className = "cell-output";
          outputArea.style.display = "block";
        } else {
          outputArea.style.display = "none";
        }
      }

      // Render display output (markdown, images, charts).
      const displays = result.displays ?? [];
      if (displays.length > 0) {
        this.renderDisplays(displayArea, displays);
      }

      void this.refreshStaleMarkers();
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      status.textContent = "Network error: " + msg;
      status.className = "status error";
    }
  }

  private renderDisplays(area: HTMLDivElement, displays: Array<{ mime: string; data: string }>): void {
    area.textContent = "";
    area.style.display = "block";

    for (const d of displays) {
      if (d.mime === "text/markdown" || d.mime === "text/plain") {
        const pre = document.createElement("pre");
        pre.className = "display-text";
        pre.textContent = d.data;
        area.appendChild(pre);
      } else if (d.mime.startsWith("image/svg+xml")) {
        // Render SVG chart inline (trusted, tool-generated).
        const container = document.createElement("div");
        container.className = "display-chart";
        container.innerHTML = d.data;
        area.appendChild(container);
        // Initialize chart interactivity if chart script is loaded.
        this.initChartInteraction();
      } else if (d.mime.startsWith("image/")) {
        // Render image as base64 data URI.
        const img = document.createElement("img");
        img.src = "data:" + d.mime + ";base64," + d.data;
        img.className = "display-image";
        img.alt = "Cell output";
        area.appendChild(img);
      } else {
        // Unknown MIME type — show as text.
        const pre = document.createElement("pre");
        pre.className = "display-text";
        pre.textContent = "[" + d.mime + "] " + d.data;
        area.appendChild(pre);
      }
    }
  }

  private initChartInteraction(): void {
    // Load chart script dynamically if not already loaded.
    if (!document.querySelector("script[src='/api/chart.js']")) {
      const script = document.createElement("script");
      script.src = "/api/chart.js";
      script.onload = (): void => {
        // Re-trigger chart initialization after script loads.
        const event = new Event("DOMContentLoaded");
        document.dispatchEvent(event);
      };
      document.head.appendChild(script);
    }
  }

  private displayErrors(
    outputArea: HTMLPreElement,
    errors: Array<{ message: string; cell: string; file: string; line: number; kind: string }>,
  ): void {
    outputArea.innerHTML = "";
    outputArea.className = "cell-output cell-errors";
    outputArea.style.display = "block";

    for (const err of errors) {
      const div = document.createElement("div");
      div.className = "error-entry error-" + (err.kind || "runtime");

      let prefix = "";
      if (err.kind === "compile" && err.line > 0) {
        prefix = "Line " + err.line + ": ";
      } else if (err.kind === "timeout") {
        prefix = "⏱ ";
      } else if (err.kind === "panic") {
        prefix = "⚠ ";
      }

      div.textContent = prefix + err.message;
      outputArea.appendChild(div);
    }
  }

  private async saveCell(
    section: HTMLElement,
    line: number,
    status: HTMLElement,
    heading: HTMLElement,
  ): Promise<void> {
    status.textContent = "Saving…";
    status.className = "status";

    const req: CellUpdateRequest = {
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
      const newName = req.name;
      heading.textContent = newName ? "Cell: " + newName : "Cell (line " + line + ")";
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      status.textContent = "Network error: " + msg;
      status.className = "status error";
    }
  }

  // --- Read-only blocks ---

  private buildDefsBlock(block: BlockInfo, isFirst: boolean = false, isLast: boolean = false): HTMLElement {
    const section = document.createElement("section");
    section.className = "block defs-block";
    const heading = document.createElement("h3");
    heading.appendChild(this.buildBlockActions(block, isFirst, isLast));
    heading.appendChild(document.createTextNode("Definitions (line " + block.line + ")"));
    section.appendChild(heading);
    const pre = document.createElement("pre");
    pre.textContent = block.body;
    section.appendChild(pre);
    return section;
  }

  private buildOutputBlock(block: BlockInfo, isFirst: boolean = false, isLast: boolean = false): HTMLElement {
    const section = document.createElement("section");
    section.className = "block output-block";
    const heading = document.createElement("h3");
    heading.appendChild(this.buildBlockActions(block, isFirst, isLast));
    heading.appendChild(document.createTextNode("Output (line " + block.line + ")"));
    section.appendChild(heading);
    const pre = document.createElement("pre");
    pre.textContent = block.body;
    section.appendChild(pre);
    return section;
  }

  // --- Helpers ---

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
    ta.rows = Math.max(4, (value ?? "").split("\n").length + 1);
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

  private showError(msg: string): void {
    const p = document.createElement("p");
    p.className = "status error";
    p.textContent = msg;
    this.container.appendChild(p);
  }
}

document.addEventListener("DOMContentLoaded", () => {
  const container = document.getElementById("notebook");
  if (!container) return;
  const editor = new NotebookEditor(container);
  void editor.load();
});
