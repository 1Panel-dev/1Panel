import { ElMessage } from 'element-plus';

export const MsgSuccess = (message) => {
    ElMessage.success({
        message: message,
        type: 'success',
        showClose: true,
        duration: 3000,
    });
};

export const MsgInfo = (message) => {
    ElMessage.info({
        message: message,
        type: 'info',
        showClose: true,
        duration: 3000,
    });
};

export const MsgWarning = (message) => {
    ElMessage.warning({
        message: message,
        type: 'warning',
        showClose: true,
        duration: 3000,
    });
};

export const MsgError = (message) => {
    ElMessage.error({
        message: message,
        type: 'error',
        showClose: true,
        duration: 3000,
    });
};
