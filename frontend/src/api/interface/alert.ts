import { CommonModel, ReqPage } from '@/api/interface';

export namespace Alert {
    export interface AlertInfo extends CommonModel {
        type: string;
        subType: string;
        cycle: number;
        count: number;
        interval: number;
        method: string;
        title: string;
        project: string;
        status: string;
        sendCount: number;
        sendMethod: string[];
    }

    export interface AlertDetail {
        type: string;
        licenseId: string;
        title: string;
        project: string;
        method: string;
        params: string;
    }

    export interface AlertUpdateStatusReq {
        id: number;
        status: string;
    }

    export interface AlertLog extends CommonModel {
        alertId: number;
        alertDetail: AlertDetail;
        alertRule: AlertInfo;
        count: number;
        message: string;
        status: string;
        method: string;
    }

    export interface DisksDTO {
        path: string;
        type: string;
        device: string;
        total: number;
        free: number;
        used: number;
        usedPercent: number;
        inodesTotal: number;
        inodesUsed: number;
        inodesFree: number;
        inodesUsedPercent: number;
    }

    export interface AlertSearch extends ReqPage {
        type: string;
        status: string;
        method: string;
        orderBy: string;
        order: string;
    }

    export interface AlertCreateReq {
        type: string;
        cycle: number;
        count: number;
        interval: number;
        method: string;
        title: string;
        project: string;
        status: string;
        sendCount: number;
    }

    export interface AlertUpdateReq extends CommonModel {
        type: string;
        cycle: number;
        count: number;
        interval: number;
        method: string;
        title: string;
        project: string;
        status: string;
        sendCount: number;
    }

    export interface DelReq {
        id: number;
    }

    export interface AlertLogSearch extends ReqPage {
        status: string;
        count: number;
    }

    export interface AlertLogId {
        id: number;
    }

    export interface ClamsDTO {
        id: number;
        name: string;
        status: string;
        path: string;
        createdAt: string;
        updatedAt: string;
    }

    export interface CronJobDTO {
        id: number;
        name: string;
        type: string;
        status: string;
        createdAt: string;
        updatedAt: string;
    }

    export interface CronJobReq {
        name: string;
        type: string;
        status: string;
    }

    export interface AlertSmsReq {
        licenseId: string;
    }

    export interface AlertSmsDTO {
        name: string;
        total: number;
        used: number;
    }

    export interface AlertConfigInfo extends CommonModel {
        type: string;
        title: string;
        config: string;
        status: string;
    }

    export interface AlertConfigUpdateReq {
        id: number;
        type: string;
        title: string;
        config: string;
        status: string;
    }

    export interface AlertConfigTest {
        port: number;
        host: string;
        sender: string;
        password: string;
        displayName: string;
        encryption: string;
        recipient: string;
    }

    export interface CommonConfig {
        isOffline?: string;
        alertDailyNum?: number;
        alertSendTimeRange?: string;
    }
}
