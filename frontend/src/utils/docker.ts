import { isJson } from './util';

export function formatImageStdout(stdout: string) {
    let lines = stdout.split('\r\n');
    for (let i = 0; i < lines.length; i++) {
        if (isJson(lines[i])) {
            const data = JSON.parse(lines[i]);
            if (data.id) {
                lines[i] = data.id + ': ' + data.status;
            } else {
                lines[i] = data.status;
            }
            if (data.progress) {
                lines[i] = lines[i] + data.progress;
            }
        }
    }
    return lines.join('\r\n');
}
