import type { Directive, DirectiveBinding } from 'vue';

const integerInput: Directive = {
    mounted(el: HTMLElement, binding: DirectiveBinding) {
        const { value } = binding;
        el.addEventListener('input', (event: Event) => {
            const inputElement = event.target as HTMLInputElement;
            let inputValue = inputElement.value;
            inputValue = inputValue.replace(/\..*/, '');
            if (value?.min !== undefined && Number(inputValue) < value.min) {
                inputValue = value.min.toString();
            }
            if (value?.max !== undefined && Number(inputValue) > value.max) {
                inputValue = value.max.toString();
            }
            inputElement.value = inputValue;
            const inputEvent = new Event('input', { bubbles: true });
            inputElement.dispatchEvent(inputEvent);
        });
    },
};

export default integerInput;
