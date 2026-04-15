<template>
    <el-form label-position="top">
        <el-form-item>
            <el-button type="primary" @click="emit('open-terminal')">
                {{ t('aiTools.agents.scanConnect') }}
            </el-button>
        </el-form-item>
    </el-form>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const emit = defineEmits(['open-terminal']);

const load = async () => {};

defineExpose({
    load,
});
</script>
