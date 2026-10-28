<template>
    <el-form-item :label="t('aiTools.agents.groupSessionScope')" prop="groupSessionScope">
        <el-select :model-value="groupSessionScope" @update:model-value="emit('update:groupSessionScope', $event)">
            <el-option :label="t('aiTools.agents.groupSessionScopeGroup')" value="group" />
            <el-option :label="t('aiTools.agents.groupSessionScopeGroupSender')" value="group_sender" />
            <el-option :label="t('aiTools.agents.feishuSessionTopic')" value="group_topic" />
            <el-option :label="t('aiTools.agents.feishuSessionTopicSender')" value="group_topic_sender" />
        </el-select>
    </el-form-item>
    <el-form-item :label="t('aiTools.agents.streaming')">
        <el-switch :model-value="streaming" @update:model-value="emit('update:streaming', Boolean($event))" />
    </el-form-item>
    <el-form-item :label="t('aiTools.agents.requireMention')" prop="requireMention">
        <el-select :model-value="requireMention" @update:model-value="emit('update:requireMention', $event)">
            <el-option :label="t('aiTools.agents.requireMentionTrue')" value="true" />
            <el-option :label="t('aiTools.agents.requireMentionFalse')" value="false" />
            <el-option v-if="allowInherit" :label="t('aiTools.agents.requireMentionOpen')" value="open" />
        </el-select>
    </el-form-item>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
defineProps<{ streaming: boolean; groupSessionScope?: string; requireMention: string; allowInherit?: boolean }>();
const emit = defineEmits<{
    'update:streaming': [value: boolean];
    'update:groupSessionScope': [value: string];
    'update:requireMention': [value: string];
}>();
</script>
