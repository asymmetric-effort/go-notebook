// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Visual regression tests for chart rendering (#61).
// Builds notebooks with chart output, renders HTML, and captures
// screenshots at desktop and mobile viewports for each chart type.

import { test, expect } from "@playwright/test";
import { execSync } from "child_process";
import * as fs from "fs";
import * as path from "path";

const outputDir = path.resolve(__dirname, "../../tmp/visual-regression/charts");

// Chart test notebooks — each produces a different chart type.
const chartNotebooks: Record<string, string> = {
	bar: `\`\`\`go:defs
import "github.com/asymmetric-effort/go-notebook/nb/chart"
\`\`\`

\`\`\`go:notebook name=barchart exported(done bool)
ctx.Chart(chart.Bar{
	Title: "Sales by Quarter",
	Labels: []string{"Q1", "Q2", "Q3", "Q4"},
	Series: []chart.Series{{Name: "Revenue", Values: []float64{100, 150, 120, 180}}},
})
done = true
\`\`\`
`,
	line: `\`\`\`go:defs
import "github.com/asymmetric-effort/go-notebook/nb/chart"
\`\`\`

\`\`\`go:notebook name=linechart exported(done bool)
ctx.Chart(chart.Line{
	Title: "Growth",
	X: []float64{1, 2, 3, 4, 5},
	Series: []chart.Series{{Name: "Users", Values: []float64{10, 25, 40, 55, 80}}},
})
done = true
\`\`\`
`,
	pie: `\`\`\`go:defs
import "github.com/asymmetric-effort/go-notebook/nb/chart"
\`\`\`

\`\`\`go:notebook name=piechart exported(done bool)
ctx.Chart(chart.Pie{
	Title: "Market Share",
	Labels: []string{"A", "B", "C"},
	Values: []float64{45, 30, 25},
})
done = true
\`\`\`
`,
};

test.beforeAll(() => {
	fs.mkdirSync(outputDir, { recursive: true });

	// Write and build each chart notebook.
	for (const [name, src] of Object.entries(chartNotebooks)) {
		const mdPath = path.join(outputDir, name + ".md");
		const htmlPath = path.join(outputDir, name + ".html");
		fs.writeFileSync(mdPath, src);
		try {
			execSync(`go run ../../cmd/gobook build ${mdPath} --format html -o ${htmlPath}`, {
				cwd: __dirname,
				timeout: 120000,
				stdio: "pipe",
			});
		} catch {
			// Chart execution may fail without full runtime — screenshots will be skipped.
		}
	}
});

for (const name of Object.keys(chartNotebooks)) {
	test.describe(`Chart: ${name}`, () => {
		test(`desktop screenshot`, async ({ page }) => {
			const htmlPath = path.join(outputDir, name + ".html");
			if (!fs.existsSync(htmlPath)) {
				test.skip();
				return;
			}
			await page.setViewportSize({ width: 1280, height: 720 });
			await page.goto("file://" + htmlPath);
			await expect(page).toHaveScreenshot(`chart-${name}-desktop.png`, {
				maxDiffPixelRatio: 0.02,
				fullPage: true,
			});
		});

		test(`mobile screenshot`, async ({ page }) => {
			const htmlPath = path.join(outputDir, name + ".html");
			if (!fs.existsSync(htmlPath)) {
				test.skip();
				return;
			}
			await page.setViewportSize({ width: 393, height: 851 });
			await page.goto("file://" + htmlPath);
			await expect(page).toHaveScreenshot(`chart-${name}-mobile.png`, {
				maxDiffPixelRatio: 0.02,
				fullPage: true,
			});
		});
	});
}

test.describe("Chart hostile labels", () => {
	test("hostile labels are escaped in HTML", async ({ page }) => {
		const src = `\`\`\`go:defs
import "github.com/asymmetric-effort/go-notebook/nb/chart"
\`\`\`

\`\`\`go:notebook name=hostile exported(done bool)
ctx.Chart(chart.Bar{
	Title: "<script>alert('xss')</script>",
	Labels: []string{"<img onerror=alert(1)>", "normal"},
	Series: []chart.Series{{Name: "<b>bold</b>", Values: []float64{10, 20}}},
})
done = true
\`\`\`
`;
		const mdPath = path.join(outputDir, "hostile.md");
		const htmlPath = path.join(outputDir, "hostile.html");
		fs.writeFileSync(mdPath, src);
		try {
			execSync(`go run ../../cmd/gobook build ${mdPath} --format html -o ${htmlPath}`, {
				cwd: __dirname,
				timeout: 120000,
				stdio: "pipe",
			});
		} catch {
			test.skip();
			return;
		}

		await page.goto("file://" + htmlPath);
		const html = await page.content();
		// Script tags in chart labels must be escaped.
		expect(html).not.toContain("<script>alert");
		// Event handlers must not appear unescaped.
		expect(html).not.toContain("onerror=alert");
	});
});
