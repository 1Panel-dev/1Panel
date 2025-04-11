import { Runtime } from '@/api/interface/runtime';

export function disabledButton(row: Runtime.Runtime, type: string): boolean {
    switch (type) {
        case 'stop':
            return (
                row.status === 'Recreating' ||
                row.status === 'Stopped' ||
                row.status === 'Building' ||
                row.resource == 'local'
            );
        case 'start':
            return (
                row.status === 'Starting' ||
                row.status === 'Recreating' ||
                row.status === 'Running' ||
                row.status === 'Building' ||
                row.resource == 'local'
            );
        case 'restart':
            return row.status === 'Recreating' || row.status === 'Building' || row.resource == 'local';
        case 'edit':
            return row.status === 'Recreating' || row.status === 'Building';
        case 'extension':
        case 'config':
            return row.status != 'Running';
        default:
            return false;
    }
}
