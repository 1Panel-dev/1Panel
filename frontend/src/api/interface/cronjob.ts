export namespace Cronjob {
    export interface CronjobInfo {
        id: number;
        name: string;
        type: string;
        specType: string;
        week: number;
        day: number;
        hour: number;
        minute: number;

        script: string;
        website: string;
        exclusionRules: string;
        database: string;
        url: string;
        targetDirID: number;
        targetDir: string;
        retainCopies: number;
        status: string;
    }
    export interface CronjobCreate {
        name: string;
        type: string;
        specType: string;
        week: number;
        day: number;
        hour: number;
        minute: number;

        script: string;
        website: string;
        exclusionRules: string;
        database: string;
        url: string;
        targetDirID: number;
        retainCopies: number;
    }
    export interface CronjobUpdate {
        id: number;
        specType: string;
        week: number;
        day: number;
        hour: number;
        minute: number;

        script: string;
        website: string;
        exclusionRules: string;
        database: string;
        url: string;
        targetDirID: number;
        retainCopies: number;
        status: string;
    }
}
