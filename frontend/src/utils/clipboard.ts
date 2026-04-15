import i18n from '@/lang';
import useClipboard from 'vue-clipboard3';

import { MsgError, MsgSuccess } from '@/utils/message';

const { toClipboard } = useClipboard();

export async function copyText(content: string) {
    try {
        await toClipboard(content);
        MsgSuccess(i18n.global.t('commons.msg.copySuccess'));
    } catch (e) {
        MsgError(i18n.global.t('commons.msg.copyFailed'));
    }
}
