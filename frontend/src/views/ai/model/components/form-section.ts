import type { FormInstance } from 'element-plus';

export const expandInvalidFormSection = (form: FormInstance, prop: string | string[], valid: boolean) => {
    if (valid) return;
    const section = form?.getField(prop)?.$el?.closest<HTMLDetailsElement>('details.form-section');
    if (section) section.open = true;
};
