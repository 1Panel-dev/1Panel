import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { OperateRuntime, updateRemark } from '@/api/modules/runtime';
import { Ref } from 'vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Runtime } from '@/api/interface/runtime';

export const operateRuntime = async (operate: string, ID: number, loading: Ref<boolean>, search: () => void) => {
    try {
        const action = await ElMessageBox.confirm(
            i18n.global.t('runtime.operatorHelper', [i18n.global.t('commons.operate.' + operate)]),
            i18n.global.t('commons.operate.' + operate),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        );

        if (action === 'confirm') {
            loading.value = true;
            await OperateRuntime({ operate: operate, ID: ID });
            search();
        }
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

export const updateRuntimeRemark = async (row: Runtime.Runtime, bulr: Function) => {
    bulr();
    if (row.remark && row.remark.length > 128) {
        MsgError(i18n.global.t('commons.rule.length128Err'));
        return;
    }
    try {
        await updateRemark({
            id: row.id,
            remark: row.remark,
        }).then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        });
    } catch (error) {}
};
