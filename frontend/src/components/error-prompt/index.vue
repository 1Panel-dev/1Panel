<template>
    <div class="msg">
        <span>{{ props.title }}</span>
    </div>
</template>

<script lang="ts" setup>
defineOptions({ name: 'ErrPrompt' });
const props = defineProps({
    title: {
        type: String,
        default: '',
    },
});
</script>

<style lang="scss" scoped>
.msg {
    color: #ff0000;
    line-height: 30px;
}
</style>
