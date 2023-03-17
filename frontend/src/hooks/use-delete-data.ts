import { ElMessageBox } from 'element-plus';
import { HandleData } from './interface';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

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
        ElMessageBox.confirm(i18n.global.t(`${message}`), i18n.global.t('commons.msg.deleteTitle'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            closeOnClickModal: false,
            closeOnPressEscape: false,
            showClose: false,
            type: confirmType,
            draggable: true,
            beforeClose: async (action, instance, done) => {
                if (action === 'confirm') {
                    instance.confirmButtonLoading = true;
                    instance.cancelButtonLoading = true;

                    await api(params)
                        .then((res) => {
                            done();
                            if (!res) return reject(false);
                            resolve(true);
                            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                        })
                        .finally(() => {
                            instance.confirmButtonLoading = false;
                            instance.cancelButtonLoading = false;
                        });
                } else {
                    done();
                }
            },
        })
            .then(() => {})
            .catch(() => {});
    });
};
