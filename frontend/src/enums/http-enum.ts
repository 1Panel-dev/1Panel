export enum ResultEnum {
    SUCCESS = 200,
    ERRIP = 310,
    ERRDOMAIN = 311,
    UNSAFETY = 312,
    EXPIRED = 313,

    ERROR = 500,
    OVERDUE = 401,
    FORBIDDEN = 403,
    NOTFOUND = 404,
    ERRAUTH = 406,
    ERRGLOBALLOADDING = 407,
    ERRXPACK = 410,
    NodeUnBind = 411,
    TIMEOUT = 20000,
    TYPE = 'success',
}

export enum TimeoutEnum {
    T_40S = 40000,
    T_60S = 60000,
    T_3M = 180000,
    T_5M = 300000,
    T_10M = 600000,
    T_1H = 3600000,
    T_1D = 86400000,
}
