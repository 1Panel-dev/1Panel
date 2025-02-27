import { Runtime } from '@/api/interface/runtime';

export function disabledButton(row: Runtime.Runtime, type: string): boolean {
    switch (type) {
        case 'stop':
            return row.status === 'Recreating' || row.status === 'Stopped' || row.status === 'Building';
        case 'start':
            return (
                row.status === 'Starting' ||
                row.status === 'Recreating' ||
                row.status === 'Running' ||
                row.status === 'Building'
            );
        case 'restart':
            return row.status === 'Recreating' || row.status === 'Building';
        case 'edit':
            return row.status === 'Recreating' || row.status === 'Building';
        case 'extension':
        case 'config':
            return row.status != 'Running';
        default:
            return false;
    }
}
