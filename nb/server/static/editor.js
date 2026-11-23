"use strict";
// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.
class NotebookEditor {
    container;
    constructor(container) {
        this.container = container;
    }
    async load() {
        const resp = await fetch("/api/blocks");
        if (!resp.ok) {
            const err = await resp.json();
            this.showError("Failed to load notebook: " + err.error);
            return;
        }
        const blocks = await resp.json();
        this.render(blocks);
    }
    render(blocks) {
        this.container.innerHTML = "";
        if (blocks.length === 0) {
            const p = document.createElement("p");
            p.textContent = "Empty notebook.";
            this.container.appendChild(p);
            return;
        }
        for (const block of blocks) {
            switch (block.kind) {
                case "prose":
                    this.container.appendChild(this.buildProseForm(block));
                    break;
                case "cell":
                    this.container.appendChild(this.buildCellForm(block));
                    break;
                case "defs":
                    this.container.appendChild(this.buildDefsBlock(block));
                    break;
                case "output":
                    this.container.appendChild(this.buildOutputBlock(block));
                    break;
            }
        }
    }
    // --- Prose editing ---
    buildProseForm(block) {
        const section = document.createElement("section");
        section.className = "block prose-block";
        const heading = document.createElement("h3");
        heading.textContent = "Prose (line " + block.line + ")";
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
    async saveProse(line, ta, status) {
        status.textContent = "Saving…";
        status.className = "status";
        const req = {
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
                const err = await resp.json();
                status.textContent = "Error: " + err.error;
                status.className = "status error";
                return;
            }
            status.textContent = "Saved.";
            status.className = "status ok";
        }
        catch (e) {
            const msg = e instanceof Error ? e.message : String(e);
            status.textContent = "Network error: " + msg;
            status.className = "status error";
        }
    }
    // --- Cell editing ---
    buildCellForm(block) {
        const section = document.createElement("section");
        section.className = "block cell-block";
        section.dataset["line"] = String(block.line);
        const heading = document.createElement("h2");
        heading.textContent = block.name ? "Cell: " + block.name : "Cell (line " + block.line + ")";
        section.appendChild(heading);
        const status = document.createElement("p");
        status.className = "status";
        section.appendChild(status);
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
        const saveBtn = document.createElement("button");
        saveBtn.type = "button";
        saveBtn.textContent = "Save Cell";
        saveBtn.addEventListener("click", () => {
            void this.saveCell(section, block.line, status, heading);
        });
        section.appendChild(saveBtn);
        return section;
    }
    async saveCell(section, line, status, heading) {
        status.textContent = "Saving…";
        status.className = "status";
        const req = {
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
                const err = await resp.json();
                status.textContent = "Error: " + err.error;
                status.className = "status error";
                return;
            }
            status.textContent = "Saved.";
            status.className = "status ok";
            const newName = req.name;
            heading.textContent = newName ? "Cell: " + newName : "Cell (line " + line + ")";
        }
        catch (e) {
            const msg = e instanceof Error ? e.message : String(e);
            status.textContent = "Network error: " + msg;
            status.className = "status error";
        }
    }
    // --- Read-only blocks ---
    buildDefsBlock(block) {
        const section = document.createElement("section");
        section.className = "block defs-block";
        const heading = document.createElement("h3");
        heading.textContent = "Definitions (line " + block.line + ")";
        section.appendChild(heading);
        const pre = document.createElement("pre");
        pre.textContent = block.body;
        section.appendChild(pre);
        return section;
    }
    buildOutputBlock(block) {
        const section = document.createElement("section");
        section.className = "block output-block";
        const heading = document.createElement("h3");
        heading.textContent = "Output (line " + block.line + ")";
        section.appendChild(heading);
        const pre = document.createElement("pre");
        pre.textContent = block.body;
        section.appendChild(pre);
        return section;
    }
    // --- Helpers ---
    labeledInput(label, name, value) {
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
    labeledTextarea(label, name, value, monospace) {
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
    labeledSelect(label, name, options, selected) {
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
    parseReads(raw) {
        if (!raw.trim())
            return [];
        return raw.split(",").map(s => s.trim()).filter(s => s.length > 0).map(s => {
            const idx = s.lastIndexOf(":");
            if (idx < 0)
                return { name: s, mode: "embed" };
            const mode = s.slice(idx + 1).trim();
            return {
                name: s.slice(0, idx).trim(),
                mode: (mode === "runtime" ? "runtime" : "embed"),
            };
        });
    }
    parseCSV(raw) {
        if (!raw.trim())
            return [];
        return raw.split(",").map(s => s.trim()).filter(s => s.length > 0);
    }
    getValue(section, name) {
        const el = section.querySelector(`[name="${name}"]`);
        return el ? el.value : "";
    }
    showError(msg) {
        const p = document.createElement("p");
        p.className = "status error";
        p.textContent = msg;
        this.container.appendChild(p);
    }
}
document.addEventListener("DOMContentLoaded", () => {
    const container = document.getElementById("notebook");
    if (!container)
        return;
    const editor = new NotebookEditor(container);
    void editor.load();
});
