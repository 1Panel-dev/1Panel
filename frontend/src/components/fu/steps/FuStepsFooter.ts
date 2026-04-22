import { computed, defineComponent, h, inject, ref } from 'vue';

import FuStepButton from './FuStepButton.vue';
import type { Stepper } from './Stepper';

export default defineComponent({
    name: 'FuStepsFooter',
    emits: ['stepperFn'],
    setup(_props, { emit }) {
        const stepper = inject<Stepper>('stepper');
        const disabledButton = ref(false);

        const isFirst = computed(() => {
            return stepper?.isFirst(stepper.index) ?? true;
        });

        const isLast = computed(() => {
            return stepper?.isLast(stepper.index) ?? true;
        });

        const showCancel = computed(() => {
            return stepper?.showCancel !== false;
        });

        const disabled = computed(() => {
            return Boolean(stepper?.isLoading || disabledButton.value);
        });

        const clickHandle = (fnName: string) => {
            if (!stepper) {
                return;
            }
            const fn = (stepper as any)[fnName];
            if (typeof fn === 'function') {
                fn.call(stepper);
            } else {
                emit('stepperFn', fnName);
            }
            disabledButton.value = true;
            setTimeout(() => {
                disabledButton.value = false;
            }, 500);
        };

        const renderButton = (value: string) => {
            if (!stepper) {
                return null;
            }
            return h(
                FuStepButton,
                {
                    disabled: disabled.value,
                    size: stepper.buttonSize,
                    onClick: () => clickHandle(value),
                },
                () => (stepper as any)[`${value}ButtonText`],
            );
        };

        return () =>
            h('div', { class: `fu-steps__footer--${stepper?.footerAlign ?? 'flex'}` }, [
                h(
                    'div',
                    {
                        class: 'fu-steps__footer--block',
                        style: 'margin-right: 10px',
                    },
                    [showCancel.value ? renderButton('onCancel') : null],
                ),
                h('div', { class: 'fu-steps__footer--block' }, [
                    !isFirst.value ? renderButton('prev') : null,
                    isLast.value ? renderButton('onFinish') : renderButton('next'),
                ]),
            ]);
    },
});
