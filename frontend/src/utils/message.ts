import { ElMessage } from 'element-plus';

let messageDom: any = null;
const messageTypes: Array<string> = ['success', 'error', 'warning', 'info'];
const Message: any = (options) => {
    if (messageDom) messageDom.close();
    messageDom = ElMessage(options);
};
messageTypes.forEach((type) => {
    Message[type] = (options) => {
        if (typeof options === 'string') options = { message: options };
        options.type = type;
        return Message(options);
    };
});

export const MsgSuccess = (message) => {
    Message.success({
        message: message,
        type: 'success',
        showClose: true,
        duration: 3000,
    });
};

export const MsgInfo = (message) => {
    Message.info({
        message: message,
        type: 'info',
        showClose: true,
        duration: 3000,
    });
};

export const MsgWarning = (message) => {
    Message.warning({
        message: message,
        type: 'warning',
        showClose: true,
        duration: 3000,
    });
};

export const MsgError = (message) => {
    Message.error({
        message: message,
        type: 'error',
        showClose: true,
        duration: 3000,
    });
};
