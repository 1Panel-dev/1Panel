import { defineComponent, h, ref } from 'vue';

import FuHorizontalSteps from './FuHorizontalSteps';
import FuVerticalSteps from './FuVerticalSteps';

export default defineComponent({
    name: 'FuSteps',
    props: {
        direction: {
            type: String,
            default: 'horizontal',
        },
    },
    emits: ['change'],
    setup(props, { attrs, slots, emit, expose }) {
        const stepsRef = ref<{
            next?: () => unknown;
            prev?: () => unknown;
            active?: (index: number) => unknown;
        } | null>(null);

        const next = () => stepsRef.value?.next();
        const prev = () => stepsRef.value?.prev();
        const active = (index: number) => stepsRef.value?.active(index);

        const handleChange = (payload: any) => {
            emit('change', payload);
        };

        expose({
            next,
            prev,
            active,
        });

        if (props.direction === 'vertical') {
            return () => h(FuVerticalSteps, { ref: stepsRef, onChange: handleChange, ...attrs }, slots);
        }

        return () => h(FuHorizontalSteps, { ref: stepsRef, onChange: handleChange, ...attrs }, slots);
    },
});
