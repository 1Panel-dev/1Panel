import { defineComponent, h, provide, ref, watch } from 'vue';

import { flattenVNodes, getVNodeComponentName } from '../shared';
import FuStepsFooter from './FuStepsFooter';
import FuVerticalNavigation from './FuVerticalNavigation.vue';
import { Step, Stepper } from './Stepper';

export default defineComponent({
    name: 'FuVerticalSteps',
    emits: ['change', 'next', 'prev', 'onCancel', 'onFinish'],
    setup(_props, { attrs, slots, emit, expose }) {
        const stepper = ref(new Stepper());
        stepper.value.activeSet.add(0);

        watch(
            () => stepper.value.index,
            (value) => {
                emit('change', stepper.value.steps[value]);
            },
        );

        const active = (index: number) => stepper.value.active(index);
        const disable = (index: number) => !stepper.value.isActive(index);
        const next = () => stepper.value.next();
        const prev = () => stepper.value.prev();
        const emitStepperFn = (name: 'next' | 'prev' | 'onCancel' | 'onFinish') => emit(name);

        provide('stepper', stepper.value);

        expose({ next, prev, active });

        return () => {
            const stepNodes = flattenVNodes(slots.default?.() || []).filter(
                (node) => getVNodeComponentName(node) === 'FuStep',
            );
            const steps = stepNodes.map((node, index) => new Step({ index, ...((node.props || {}) as object) }));
            const currentNode = stepNodes.find((_node, index) => stepper.value.isCurrent(index));

            Object.assign(stepper.value, attrs);
            stepper.value.steps = steps;

            return h('div', { class: ['fu-steps', 'fu-steps--vertical'] }, [
                h(
                    FuVerticalNavigation,
                    {
                        stepper: stepper.value,
                        steps,
                        disable,
                        onActive: active,
                    },
                    () => currentNode,
                ),
                h(
                    'div',
                    { class: 'fu-steps__footer' },
                    slots.footer?.() || h(FuStepsFooter, { onStepperFn: emitStepperFn }),
                ),
            ]);
        };
    },
});
