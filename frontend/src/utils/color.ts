export function adjustColorToRGBA(color: string, percent: number, opacity: number): string {
    let r = 0,
        g = 0,
        b = 0,
        a = opacity;

    color = color.trim();

    if (color.startsWith('#')) {
        if (color.length === 4) {
            r = parseInt(color[1] + color[1], 16);
            g = parseInt(color[2] + color[2], 16);
            b = parseInt(color[3] + color[3], 16);
        } else if (color.length === 7) {
            r = parseInt(color.slice(1, 3), 16);
            g = parseInt(color.slice(3, 5), 16);
            b = parseInt(color.slice(5, 7), 16);
        } else {
            return color;
        }
    } else if (color.startsWith('rgb')) {
        const result = color.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*([0-9.]+))?\)/);
        if (!result) return color;
        r = parseInt(result[1], 10);
        g = parseInt(result[2], 10);
        b = parseInt(result[3], 10);
        if (result[4] !== undefined) {
            a = parseFloat(result[4]);
        }
    } else {
        return color;
    }

    r = Math.min(255, Math.max(0, Math.round(r * (1 + percent / 100))));
    g = Math.min(255, Math.max(0, Math.round(g * (1 + percent / 100))));
    b = Math.min(255, Math.max(0, Math.round(b * (1 + percent / 100))));
    a = Math.min(1, Math.max(0, opacity / 100));

    return `rgba(${r}, ${g}, ${b}, ${a})`;
}
