import i18n from '@/lang';
import { FormItemRule } from 'element-plus';

interface CommonRule {
    required: FormItemRule;
    name: FormItemRule;
}

export const Rules: CommonRule = {
    required: {
        required: true,
        message: i18n.global.t('commons.rule.required'),
        trigger: 'blur',
    },
    name: {
        type: 'regexp',
        min: 1,
        max: 30,
        message: i18n.global.t('commons.rule.commonName'),
        trigger: 'blur',
        pattern: '/^[a-zA-Z0-9\u4e00-\u9fa5]{1}[a-zA-Z0-9_.\u4e00-\u9fa5-]{0,30}$/',
    },
};
