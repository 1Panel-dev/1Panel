<template>
    <el-card class="router_card">
        <el-radio-group v-model="activeName" @change="handleChange">
            <el-radio-button
                class="router_card_button"
                :label="button.label"
                v-for="(button, index) in buttonArray"
                size="large"
                :key="index"
            >
                <el-badge :value="button.count" v-if="button.count" is-dot>
                    <span>{{ button.label }}</span>
                </el-badge>
            </el-radio-button>
        </el-radio-group>
        <slot name="route-button"></slot>
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
    }

    .el-radio-button__original-radio:checked + .el-radio-button__inner {
        color: $primary-color;
        border-color: $primary-color !important;
        border-radius: 4px;
    }
}
</style>
