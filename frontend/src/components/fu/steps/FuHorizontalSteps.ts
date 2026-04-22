import { computed, defineComponent, h, provide, ref, Transition, watch } from 'vue';

import { flattenVNodes, getVNodeComponentName } from '../shared';
import FuHorizontalNavigation from './FuHorizontalNavigation.vue';
import FuStepsFooter from './FuStepsFooter';
import { Step, Stepper } from './Stepper';

export default defineComponent({
    name: 'FuHorizontalSteps',
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

        const heightStyle = computed(() => {
            const height = Number.parseInt(String(stepper.value.height ?? ''), 10);
            return Number.isFinite(height) ? { height: `${height}px` } : { height: 'auto' };
        });

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

            Object.assign(stepper.value, attrs);
            stepper.value.steps = steps;

            return h('div', { class: ['fu-steps', 'fu-steps--horizontal'] }, [
                h(FuHorizontalNavigation, {
                    stepper: stepper.value,
                    steps,
                    disable,
                    onActive: active,
                }),
                h('div', { class: 'fu-steps__wrapper' }, [
                    h(
                        'div',
                        { class: 'fu-steps__container', style: heightStyle.value },
                        h(Transition, { name: 'carousel', mode: 'out-in' }, () =>
                            stepNodes.map((node, index) =>
                                stepper.value.index === index ? h(node, { key: index }) : null,
                            ),
                        ),
                    ),
                ]),
                h(
                    'div',
                    { class: 'fu-steps__footer' },
                    slots.footer?.() || h(FuStepsFooter, { onStepperFn: emitStepperFn }),
                ),
            ]);
        };
    },
});
