<template>
    <el-card class="router_card">
        <el-radio-group v-model="activeName" @change="handleChange">
            <el-radio-button
                class="router_card_button"
                :label="button.label"
                v-for="(button, index) in buttonArray"
                size="large"
                :key="index"
            ></el-radio-button>
        </el-radio-group>
    </el-card>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

defineOptions({ name: 'RouterButton' });

const props = defineProps({
    buttons: {
        type: Array,
        required: true,
    },
});

const buttonArray: any = computed(() => {
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
    buttonArray.value.forEach((btn: RouterButton) => {
        if (btn.label == label) {
            if (btn.path) {
                routerToPath(btn.path);
            } else if (btn.name) {
                routerToName(btn.name);
            }
            activeName.value = btn.label;
            return;
        }
    });
};

onMounted(() => {
    const nowPath = router.currentRoute.value.path;
    if (buttonArray.value.length > 0) {
        let isPathExist = false;
        buttonArray.value.forEach((btn: RouterButton) => {
            if (btn.path == nowPath) {
                isPathExist = true;
                activeName.value = btn.label;
                return;
            }
        });
        if (!isPathExist) {
            activeName.value = buttonArray.value[0].label;
        }
    }
});
</script>

<style lang="scss">
.router_card {
    --el-card-border-radius: 8px;
    --el-card-padding: 0;
    padding: 0px;
    padding-bottom: 2px;
    padding-top: 2px;
}
.router_card_button {
    margin-left: 2px;
    .el-radio-button__inner {
        min-width: 100px;
        height: 100%;
        border: 0 !important;
    }

    .el-radio-button__original-radio:checked + .el-radio-button__inner {
        border-radius: 3px;
        color: $primary-color;
        background-color: #ffffff;
        box-shadow: 0 0 0 2px $primary-color !important;
    }

    .el-radio-button:first-child .el-radio-button__inner {
        border-radius: 3px;
        color: $primary-color;
        background-color: #ffffff;
        box-shadow: 0 0 0 2px $primary-color !important;
    }
}
</style>
