<template>
    <div class="overflow-y-auto overflow-x-hidden" :style="'height: ' + mainHeight + 'px'">
        <slot></slot>
    </div>
</template>

<script lang="ts" setup>
const props = defineProps({
    heightDiff: {
        type: Number,
        default: 0,
    },
});

let mainHeight = ref(0);

onMounted(() => {
    let heightDiff = 300;
    if (props.heightDiff) {
        heightDiff = props.heightDiff;
    }

    mainHeight.value = window.innerHeight - heightDiff;
    window.onresize = () => {
        return (() => {
            mainHeight.value = window.innerHeight - heightDiff;
        })();
    };
});

defineOptions({ name: 'MainDiv' });
</script>
