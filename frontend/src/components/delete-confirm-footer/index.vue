<template>
    <div class="flex flex-wrap items-center justify-between gap-2">
        <el-button
            class="el-col-offset-1 min-h-8 !px-0"
            type="primary"
            link
            native-type="button"
            :disabled="disabled || !confirmationText || modelValue === confirmationText"
            @click="emit('update:modelValue', confirmationText)"
        >
            {{ $t('commons.button.autoFill') }}
        </el-button>
        <div class="ms-auto">
            <slot />
        </div>
    </div>
</template>

<script setup lang="ts">
defineProps<{ modelValue: string; confirmationText: string; disabled?: boolean }>();
const emit = defineEmits<{ 'update:modelValue': [value: string] }>();
</script>
