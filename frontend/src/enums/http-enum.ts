export enum ResultEnum {
    SUCCESS = 200,
    ERR_IP = 310,
    ERR_DOMAIN = 311,
    UNSAFETY = 312,
    EXPIRED = 313,

    ERROR = 500,
    OVERDUE = 401,
    FORBIDDEN = 403,
    NOTFOUND = 404,
    ERR_AUTH = 406,
    ERR_GLOBAL_LOADING = 407,
    ERR_XPACK = 410,
    NODE_UNBIND = 411,
    ERR_RBAC = 412,
    ERR_ENTERPRISE = 413,
    TIMEOUT = 20000,
    TYPE = 'success',
}

export enum TimeoutEnum {
    T_40S = 40000,
    T_60S = 60000,
    T_3M = 180000,
    T_5M = 300000,
    T_10M = 600000,
}
