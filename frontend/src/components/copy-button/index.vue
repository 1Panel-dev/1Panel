<template>
    <el-button v-if="type == 'icon'" link @click="copyText(content)" icon="DocumentCopy" class="ml-1.5"></el-button>
    <el-button @click="copyText(content)" v-else>{{ $t('commons.button.copy') }}</el-button>
</template>

<script lang="ts" setup>
import { copyText } from '@/utils/util';
defineOptions({ name: 'CopyButton' });

defineProps({
    content: String,
    type: String,
});
</script>
