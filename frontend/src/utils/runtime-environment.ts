export interface EnvironmentEntry {
    key: string;
    value: string;
}

export const parseEnvironment = (text: string) => {
    const entries: EnvironmentEntry[] = [];
    const lines = text
        .replace(/^\uFEFF/, '')
        .replace(/\r\n?/g, '\n')
        .split('\n');
    for (let i = 0; i < lines.length; i++) {
        const line = lines[i].trimStart();
        if (!line || line.startsWith('#')) continue;
        const match = /^(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$/.exec(line);
        if (!match) return { entries: [], error: { line: i + 1, reason: 'envInvalidAssignment' } };
        const key = match[1];
        let value = match[2];
        const quote = value[0];
        if (quote === '"' || quote === "'") {
            const startLine = i + 1;
            let end = 1;
            // Quoted values may span lines; never evaluate variable or command substitutions.
            for (; ; end++) {
                if (end >= value.length) {
                    if (++i >= lines.length) {
                        return { entries: [], error: { line: startLine, reason: 'envUnclosedQuote' } };
                    }
                    value += '\n' + lines[i];
                }
                if (value[end] === '\\' && (value[end + 1] === quote || value[end + 1] === '\\')) {
                    end++;
                } else if (value[end] === quote) {
                    break;
                }
            }
            const trailing = value.slice(end + 1).trim();
            if (trailing && !trailing.startsWith('#')) {
                return { entries: [], error: { line: i + 1, reason: 'envUnexpectedText' } };
            }
            value = value.slice(1, end);
            if (quote === '"') {
                value = value.replace(/\\([nrt"\\])/g, (_, char: string) => {
                    return ({ n: '\n', r: '\r', t: '\t', '"': '"', '\\': '\\' } as Record<string, string>)[char];
                });
            } else {
                value = value.replace(/\\'/g, "'");
            }
        } else {
            value = value.startsWith('#') ? '' : value.replace(/\s+#.*$/, '').trimEnd();
        }
        const previous = entries.find((entry) => entry.key === key);
        if (previous) previous.value = value;
        else entries.push({ key, value });
    }
    return { entries, error: null };
};

export const mergeEnvironments = (current: EnvironmentEntry[], incoming: EnvironmentEntry[]) => {
    for (const entry of incoming) {
        const existing = current.find((item) => item.key === entry.key);
        if (existing) existing.value = entry.value;
        else current.push({ ...entry });
    }
};
