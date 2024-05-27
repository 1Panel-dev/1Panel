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
    TIMEOUT = 20000,
    TYPE = 'success',
}

export enum TimeoutEnum {
    T_40S = 40000,
    T_60S = 60000,
    T_5M = 300000,
    T_10M = 600000,
    T_1H = 3600000,
    T_1D = 86400000,
}
/**
 * @description：请求方法
 */
export enum RequestEnum {
    GET = 'GET',
    POST = 'POST',
    PATCH = 'PATCH',
    PUT = 'PUT',
    DELETE = 'DELETE',
}

/**
 * @description：常用的contentTyp类型
 */
export enum ContentTypeEnum {
    // json
    JSON = 'application/json;charset=UTF-8',
    // text
    TEXT = 'text/plain;charset=UTF-8',
    // form-data 一般配合qs
    FORM_URLENCODED = 'application/x-www-form-urlencoded;charset=UTF-8',
    // form-data 上传
    FORM_DATA = 'multipart/form-data;charset=UTF-8',
}
