import { computed, defineComponent, h, ref, watch, type PropType, type VNode } from 'vue';

import { flattenVNodes, getVNodeComponentName } from './shared';

interface FuStepItem {
    id: string;
    index: number;
    title: string;
    vnode: VNode;
}

const buildStepChildren = (vnode: VNode) => {
    const children = vnode.children as Record<string, (() => VNode[]) | undefined> | null;
    return children?.default?.() || [];
};

export default defineComponent({
    name: 'FuSteps',
    props: {
        active: {
            type: Number,
            default: 0,
        },
        direction: {
            type: String,
            default: 'horizontal',
        },
        space: {
            type: Number,
            default: 0,
        },
        isLoading: {
            type: Boolean,
            default: false,
        },
        finishButtonText: {
            type: String,
            default: '',
        },
        beforeLeave: {
            type: Function as PropType<
                (step: { id: string; index: number; title: string }) => boolean | Promise<boolean>
            >,
            default: undefined,
        },
    },
    emits: ['change'],
    setup(props, { emit, slots, expose }) {
        const activeIndex = ref(props.active);
        const isRunningBeforeLeave = ref(false);

        const stepItems = computed<FuStepItem[]>(() =>
            flattenVNodes(slots.default?.() || [])
                .filter((vnode) => getVNodeComponentName(vnode) === 'FuStep')
                .map((vnode, index) => {
                    const vnodeProps = (vnode.props || {}) as Record<string, any>;
                    return {
                        id: String(vnodeProps.id || index),
                        index,
                        title: String(vnodeProps.title || ''),
                        vnode,
                    };
                }),
        );

        const emitChange = () => {
            const currentStep = stepItems.value[activeIndex.value];
            if (!currentStep) {
                return;
            }
            emit('change', {
                id: currentStep.id,
                index: currentStep.index,
                title: currentStep.title,
            });
        };

        watch(
            () => props.active,
            (value) => {
                activeIndex.value = value;
            },
        );

        watch(
            stepItems,
            (items) => {
                if (items.length === 0) {
                    activeIndex.value = 0;
                    return;
                }
                if (activeIndex.value > items.length - 1) {
                    activeIndex.value = items.length - 1;
                }
                emitChange();
            },
            { immediate: true, deep: true },
        );

        const runBeforeLeave = async () => {
            if (!props.beforeLeave) {
                return true;
            }
            const currentStep = stepItems.value[activeIndex.value];
            if (!currentStep) {
                return true;
            }
            if (isRunningBeforeLeave.value) {
                return true;
            }
            isRunningBeforeLeave.value = true;
            try {
                return (
                    (await props.beforeLeave({
                        id: currentStep.id,
                        index: currentStep.index,
                        title: currentStep.title,
                    })) !== false
                );
            } finally {
                isRunningBeforeLeave.value = false;
            }
        };

        const changeTo = async (nextIndex: number, runGuard = true) => {
            if (nextIndex < 0 || nextIndex >= stepItems.value.length || nextIndex === activeIndex.value) {
                return false;
            }
            const currentIndex = activeIndex.value;
            if (runGuard) {
                const canLeave = await runBeforeLeave();
                if (activeIndex.value !== currentIndex) {
                    return true;
                }
                if (!canLeave) {
                    return false;
                }
            }
            activeIndex.value = nextIndex;
            emitChange();
            return true;
        };

        const next = async () => {
            if (isRunningBeforeLeave.value) {
                const nextIndex = Math.min(activeIndex.value + 1, stepItems.value.length - 1);
                if (nextIndex !== activeIndex.value) {
                    activeIndex.value = nextIndex;
                    emitChange();
                }
                return true;
            }
            return changeTo(activeIndex.value + 1, true);
        };

        const prev = async () => {
            return changeTo(activeIndex.value - 1, false);
        };

        expose({
            next,
            prev,
            active: activeIndex,
        });

        return () => {
            const currentStep = stepItems.value[activeIndex.value];
            return h('div', { class: ['fu-steps', `fu-steps--${props.direction}`] }, [
                h(
                    'div',
                    {
                        class: 'fu-steps__nav',
                        style: props.direction === 'vertical' && props.space ? { gap: `${props.space}px` } : undefined,
                    },
                    stepItems.value.map((step, index) =>
                        h(
                            'button',
                            {
                                key: step.id,
                                class: [
                                    'fu-steps__item',
                                    {
                                        'is-active': index === activeIndex.value,
                                        'is-finished': index < activeIndex.value,
                                    },
                                ],
                                type: 'button',
                                disabled: props.isLoading,
                                onClick: () => changeTo(step.index, step.index > activeIndex.value),
                            },
                            [
                                h('span', { class: 'fu-steps__index' }, index + 1),
                                h('span', { class: 'fu-steps__title' }, step.title),
                            ],
                        ),
                    ),
                ),
                h('div', { class: 'fu-steps__content' }, currentStep ? buildStepChildren(currentStep.vnode) : []),
                slots.footer
                    ? h(
                          'div',
                          { class: 'fu-steps__footer' },
                          slots.footer({
                              active: currentStep,
                              next,
                              prev,
                          }),
                      )
                    : null,
            ]);
        };
    },
});
