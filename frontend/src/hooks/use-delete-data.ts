import { ElMessageBox, ElMessage } from 'element-plus';
import { HandleData } from './interface';
import i18n from '@/lang';

/**
 * @description 删除操作使用
 * @param {Function} api 操作数据接口的api方法(必传)
 * @param {Object} params 携带的操作数据参数 {id,params}(必传)
 * @param {String} message 提示信息(必传)
 * @param {String} confirmType icon类型(不必传,默认为 warning)
 * @return Promise
 */
export const useDeleteData = <P = any, R = any>(
    api: (params: P) => Promise<R>,
    params: Parameters<typeof api>[0],
    message: string,
    confirmType: HandleData.MessageType = 'error',
) => {
    return new Promise((resolve, reject) => {
        ElMessageBox.confirm(i18n.global.t(`${message}`) + '?', i18n.global.t('commons.msg.deleteTitle'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: confirmType,
            draggable: true,
        }).then(async () => {
            const res = await api(params);
            if (!res) return reject(false);
            ElMessage({
                type: 'success',
                message: i18n.global.t('commons.msg.deleteSuccss'),
            });
            resolve(true);
        });
    });
};
