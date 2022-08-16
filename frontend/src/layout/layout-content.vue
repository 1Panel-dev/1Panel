<template>
    <div class="main-box">
        <div class="content-container__header" v-if="slots.header || header">
            <slot name="header">
                <back-button
                    :path="backPath"
                    :name="backName"
                    :to="backTo"
                    :header="header"
                    v-if="showBack"
                ></back-button>
                {{ header }}
            </slot>
        </div>
        <div class="content-container__toolbar" v-if="slots.toolbar">
            <slot name="toolbar"></slot>
        </div>
        <div class="content-container_form">
            <slot name="form">
                <form-button>
                    <slot name="button"></slot>
                </form-button>
            </slot>
        </div>
        <slot></slot>
    </div>
</template>

<script setup lang="ts">
import { computed, useSlots } from 'vue';
import BackButton from '@/components/back-button/index.vue';
import FormButton from './form-button.vue';
defineOptions({ name: 'LayoutContent' }); // 组件名
const slots = useSlots();
const prop = defineProps({
    header: String,
    backPath: String,
    backName: String,
    backTo: Object,
});

const showBack = computed(() => {
    const { backPath, backName, backTo } = prop;
    return backPath || backName || backTo;
});
</script>

<style lang="scss">
@use '@/styles/mixins.scss' as *;

.content-container__header {
    font-weight: 700;
    padding: 5px 0 25px;
    font-size: 18px;
}

.content-container__toolbar {
    @include flex-row(space-between, center);
    margin-bottom: 10px;
}

.content-container_form {
    text-align: -webkit-center;
    width: 60%;
    margin-left: 15%;
    .form-button {
        float: right;
    }
}
</style>
