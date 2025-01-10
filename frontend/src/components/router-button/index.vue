<template>
    <el-card class="router_card p-1 sm:p-0">
        <div class="flex w-full flex-col justify-start sm:items-center items-start sm:justify-between sm:flex-row">
            <el-radio-group v-model="activeName" @change="handleChange" class="flex-1">
                <el-radio-button
                    class="router_card_button"
                    :label="button.label"
                    :value="button.label"
                    v-for="(button, index) in buttonArray"
                    size="large"
                    :key="index"
                >
                    <el-badge :value="button.count" v-if="button.count" is-dot>
                        <span>{{ button.label }}</span>
                    </el-badge>
                </el-radio-button>
            </el-radio-group>
            <div class="flex flex-col gap-2 sm:flex-row">
                <slot name="route-button"></slot>
            </div>
        </div>
    </el-card>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

defineOptions({ name: 'RouterButton' });

const props = defineProps({
    buttons: {
        type: Array<RouterButton>,
        required: true,
    },
});

const buttonArray = computed(() => {
    return props.buttons;
});

const router = useRouter();
const activeName = ref('');
const routerToPath = (path: string) => {
    router.push({ path: path });
};
const routerToName = (name: string) => {
    router.push({ name: name });
};

const handleChange = (label: string) => {
    const btn = buttonArray.value.find((btn) => btn.label === label);
    if (!btn) return;
    if (btn.path) routerToPath(btn.path);
    else if (btn.name) routerToName(btn.name);
    activeName.value = btn.label;
};

onMounted(() => {
    if (buttonArray.value.length) {
        let isPathExist = false;
        const btn = buttonArray.value.find((btn) => {
            return router.currentRoute.value.path.startsWith(btn.path);
        });
        if (btn) {
            isPathExist = true;
            activeName.value = btn.label;
        }
        if (!isPathExist) {
            activeName.value = buttonArray.value[0].label;
        }
    }
});
</script>

<style lang="scss">
.router_card {
    --el-card-padding: 0;
    .el-card__body {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
}

.router_card_button {
    .el-radio-button__inner {
        min-width: 100px;
        height: 100%;
        background-color: var(--panel-button-active) !important;
        box-shadow: none !important;
        border: 2px solid transparent !important;
        color: var(--el-text-color-regular) !important;
    }

    .el-radio-button__original-radio:checked + .el-radio-button__inner {
        color: var(--panel-button-text-color) !important;
        background-color: var(--panel-button-bg-color) !important;
        border-color: var(--panel-color-primary) !important;
        border-radius: 4px;
    }
}
</style>
