import { useI18n } from 'vue-i18n';
export function tr(key: any) {
    const { t } = useI18n();
    key as string;
    return t(key);
}
