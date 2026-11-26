/**
 * Simple markdown-to-HTML converter.
 * Zero dependencies — handles the subset of markdown used in the project docs.
 */

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

function processInline(text: string): string {
  // Images: ![alt](src)
  text = text.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, '<img src="$2" alt="$1" />');
  // Links: [text](url)
  text = text.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>');
  // Bold: **text**
  text = text.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  // Italic: *text* (but not inside **)
  text = text.replace(/(?<!\*)\*(?!\*)(.+?)(?<!\*)\*(?!\*)/g, '<em>$1</em>');
  // Inline code: `text`
  text = text.replace(/`([^`]+)`/g, '<code>$1</code>');
  return text;
}

export function markdownToHtml(markdown: string): string {
  const lines = markdown.split('\n');
  const result: string[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    // Blank line
    if (line.trim() === '') {
      i++;
      continue;
    }

    // Fenced code block
    const codeMatch = line.match(/^```(\w*)/);
    if (codeMatch) {
      const lang = codeMatch[1];
      const codeLines: string[] = [];
      i++;
      while (i < lines.length && !lines[i].startsWith('```')) {
        codeLines.push(lines[i]);
        i++;
      }
      i++; // skip closing ```

      const cls = lang ? ` class="language-${lang}"` : '';
      result.push(`<pre><code${cls}>${escapeHtml(codeLines.join('\n'))}</code></pre>`);
      continue;
    }

    // Heading
    const headingMatch = line.match(/^(#{1,6})\s+(.+)/);
    if (headingMatch) {
      const level = headingMatch[1].length;
      result.push(`<h${level}>${processInline(escapeHtml(headingMatch[2]))}</h${level}>`);
      i++;
      continue;
    }

    // Horizontal rule
    if (/^(-{3,}|_{3,}|\*{3,})\s*$/.test(line.trim())) {
      result.push('<hr />');
      i++;
      continue;
    }

    // Blockquote
    if (line.trimStart().startsWith('> ')) {
      const quoteLines: string[] = [];
      while (i < lines.length && (lines[i].trimStart().startsWith('> ') || lines[i].trimStart().startsWith('>'))) {
        quoteLines.push(lines[i].trimStart().replace(/^>\s?/, ''));
        i++;
      }
      result.push(`<blockquote>${markdownToHtml(quoteLines.join('\n'))}</blockquote>`);
      continue;
    }

    // Table
    if (line.includes('|') && i + 1 < lines.length && /^\s*\|?\s*[-:]+[-|:\s]+\s*$/.test(lines[i + 1])) {
      const parseRow = (row: string): string[] => {
        return row.split('|').map(c => c.trim()).filter((_, idx, arr) => {
          if (idx === 0 && arr[idx] === '') return false;
          if (idx === arr.length - 1 && arr[idx] === '') return false;
          return true;
        });
      };
      const headers = parseRow(line);
      i += 2; // skip header + separator
      const rows: string[][] = [];
      while (i < lines.length && lines[i].includes('|') && lines[i].trim() !== '') {
        rows.push(parseRow(lines[i]));
        i++;
      }
      let table = '<table><thead><tr>';
      for (const h of headers) {
        table += `<th>${processInline(escapeHtml(h))}</th>`;
      }
      table += '</tr></thead><tbody>';
      for (const row of rows) {
        table += '<tr>';
        for (const cell of row) {
          table += `<td>${processInline(escapeHtml(cell))}</td>`;
        }
        table += '</tr>';
      }
      table += '</tbody></table>';
      result.push(table);
      continue;
    }

    // Unordered list
    if (/^\s*[-*]\s+/.test(line)) {
      const items: string[] = [];
      while (i < lines.length && /^\s*[-*]\s+/.test(lines[i])) {
        let itemText = lines[i].replace(/^\s*[-*]\s+/, '');
        i++;
        while (i < lines.length && lines[i].match(/^\s{2,}/) && !/^\s*[-*]\s+/.test(lines[i])) {
          itemText += ' ' + lines[i].trim();
          i++;
        }
        items.push(itemText);
      }
      result.push('<ul>' + items.map(item => `<li>${processInline(escapeHtml(item))}</li>`).join('') + '</ul>');
      continue;
    }

    // Ordered list
    if (/^\s*\d+\.\s+/.test(line)) {
      const items: string[] = [];
      while (i < lines.length && /^\s*\d+\.\s+/.test(lines[i])) {
        let itemText = lines[i].replace(/^\s*\d+\.\s+/, '');
        i++;
        while (i < lines.length && lines[i].match(/^\s{2,}/) && !/^\s*\d+\.\s+/.test(lines[i])) {
          itemText += ' ' + lines[i].trim();
          i++;
        }
        items.push(itemText);
      }
      result.push('<ol>' + items.map(item => `<li>${processInline(escapeHtml(item))}</li>`).join('') + '</ol>');
      continue;
    }

    // HTML comment — skip
    if (line.trimStart().startsWith('<!--')) {
      while (i < lines.length && !lines[i].includes('-->')) {
        i++;
      }
      i++;
      continue;
    }

    // Paragraph
    const paraLines: string[] = [];
    while (
      i < lines.length &&
      lines[i].trim() !== '' &&
      !lines[i].match(/^```/) &&
      !lines[i].match(/^#{1,6}\s/) &&
      !lines[i].match(/^\s*[-*]\s+/) &&
      !lines[i].match(/^\s*\d+\.\s+/) &&
      !lines[i].match(/^\s*>/) &&
      !lines[i].match(/^(-{3,}|_{3,}|\*{3,})\s*$/) &&
      !lines[i].match(/^\s*<!--/)
    ) {
      paraLines.push(lines[i]);
      i++;
    }
    if (paraLines.length > 0) {
      result.push(`<p>${processInline(escapeHtml(paraLines.join(' ')))}</p>`);
    }
  }

  return result.join('\n');
}
