import { defineComponent, h } from 'vue';

export default defineComponent({
    name: 'FuStep',
    props: {
        id: {
            type: String,
            default: '',
        },
        title: {
            type: String,
            default: '',
        },
    },
    setup(_props, { slots }) {
        return () => h('div', slots.default?.());
    },
});
