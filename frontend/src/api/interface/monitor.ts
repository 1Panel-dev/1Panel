export namespace Monitor {
    export interface MonitorData {
        param: string;
        date: Array<Date>;
        value: Array<any>;
    }
    export interface MonitorSearch {
        param: string;
        info: string;
        startTime: Date;
        endTime: Date;
    }
}
