<template>
    <DrawerPro
        v-model="visible"
        :header="$t('aiTools.chat.title')"
        :resource="name"
        :size="isMobile ? 'full' : '50%'"
        :auto-close="false"
        full-screen
        @close="close"
    >
        <template #content>
            <div class="model-chat">
                <div class="chat-toolbar">
                    <span class="chat-mode">
                        <el-icon><ChatDotRound /></el-icon>
                        {{ $t('aiTools.chat.title') }}
                    </span>
                    <el-button :icon="Delete" :disabled="sending || !messages.length" @click="clear">
                        {{ $t('aiTools.chat.clear') }}
                    </el-button>
                </div>
                <div class="chat-messages" role="log" aria-live="polite">
                    <div v-if="!messages.length" class="chat-empty">
                        <span class="chat-empty-icon">
                            <el-icon><ChatDotRound /></el-icon>
                        </span>
                        <h3>{{ $t('aiTools.chat.welcome') }}</h3>
                        <p>{{ $t('aiTools.chat.empty') }}</p>
                    </div>
                    <div v-else class="chat-transcript">
                        <article
                            v-for="(message, index) in messages"
                            :key="index"
                            :class="['chat-message', message.role]"
                        >
                            <div class="chat-role">{{ message.role === 'user' ? $t('aiTools.chat.you') : name }}</div>
                            <div class="chat-message-body">
                                <details v-if="message.reasoning" class="chat-reasoning">
                                    <summary>{{ $t('aiTools.chat.thinking') }}</summary>
                                    <div class="chat-text">{{ message.reasoning }}</div>
                                </details>
                                <div v-if="message.role === 'user'" class="chat-text">{{ message.content }}</div>
                                <Markdown v-else-if="message.content" :content="message.content" />
                                <el-text
                                    v-else-if="sending && index === messages.length - 1"
                                    type="info"
                                    class="chat-generating"
                                >
                                    <span class="chat-generating-dot"></span>
                                    {{ $t('aiTools.chat.waiting') }}
                                </el-text>
                            </div>
                        </article>
                    </div>
                </div>
                <div class="chat-bottom">
                    <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon />
                    <el-alert
                        v-if="remainingRounds === 0"
                        :title="$t('aiTools.chat.limitReached')"
                        type="warning"
                        :closable="false"
                        show-icon
                    />
                    <el-alert
                        class="chat-notice"
                        :title="$t('aiTools.chat.temporary')"
                        type="warning"
                        :closable="false"
                        show-icon
                    />
                    <div class="chat-composer">
                        <el-input
                            v-model="input"
                            type="textarea"
                            :autosize="{ minRows: 3, maxRows: 7 }"
                            :maxlength="262144"
                            :disabled="remainingRounds === 0"
                            :placeholder="$t('aiTools.chat.placeholder')"
                            :aria-label="$t('aiTools.chat.placeholder')"
                            @keydown="onKeydown"
                        />
                        <div class="chat-actions">
                            <div v-if="target?.models?.length > 1" class="chat-model-select">
                                <span class="chat-model-width" aria-hidden="true">{{ model }}</span>
                                <el-select
                                    v-model="model"
                                    filterable
                                    size="small"
                                    :disabled="sending || messages.length > 0"
                                    :aria-label="$t('aiTools.model.model')"
                                >
                                    <el-option
                                        v-for="option in target.models"
                                        :key="option.id"
                                        :label="option.id"
                                        :value="option.id"
                                    />
                                </el-select>
                            </div>
                            <el-text v-else truncated class="chat-composer-model" :title="model">
                                {{ model }}
                            </el-text>
                            <el-tooltip v-if="sending" :content="$t('commons.operate.stop')">
                                <el-button
                                    circle
                                    size="small"
                                    :icon="VideoPause"
                                    :aria-label="$t('commons.operate.stop')"
                                    @click="controller?.abort()"
                                />
                            </el-tooltip>
                            <el-tooltip v-else :content="$t('aiTools.chat.send')">
                                <el-button
                                    circle
                                    size="small"
                                    type="primary"
                                    :icon="Promotion"
                                    :disabled="!input.trim() || remainingRounds === 0"
                                    :aria-label="$t('aiTools.chat.send')"
                                    @click="send"
                                />
                            </el-tooltip>
                        </div>
                    </div>
                </div>
            </div>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { ChatDotRound, Delete, Promotion, VideoPause } from '@element-plus/icons-vue';
import { useRoute } from 'vue-router';
import DrawerPro from '@/components/drawer-pro/index.vue';
import Markdown from '@/components/mkdown-editor/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { chatModel, type ModelChatMessage, type ModelChatTarget } from '@/api/modules/model-chat';
import i18n from '@/lang';

const { currentNode, isMobile } = useGlobalStore();
const route = useRoute();
const visible = ref(false);
const target = ref<ModelChatTarget>();
const name = computed(() => target.value?.name || '');
const model = ref('');
const input = ref('');
const error = ref('');
const sending = ref(false);
const messages = ref<(ModelChatMessage & { reasoning?: string; complete?: boolean })[]>([]);
const remainingRounds = computed(() =>
    Math.max(0, 100 - messages.value.filter((message) => message.role === 'user' && message.complete).length),
);
let controller: AbortController | undefined;

const clear = () => {
    messages.value = [];
    error.value = '';
};
const close = () => {
    controller?.abort();
    controller = undefined;
    sending.value = false;
    visible.value = false;
    input.value = '';
    clear();
};
const open = (options: ModelChatTarget) => {
    close();
    target.value = options;
    model.value = options.model;
    visible.value = true;
};
const send = async () => {
    if (!target.value || sending.value || !input.value.trim() || remainingRounds.value === 0) return;
    const content = input.value.trim();
    const history = messages.value
        .filter((message) => message.complete)
        .map(({ role, content }) => ({ role, content }));
    const request = new AbortController();
    controller = request;
    messages.value.push({ role: 'user', content }, { role: 'assistant', content: '', reasoning: '' });
    const user = messages.value[messages.value.length - 2];
    const reply = messages.value[messages.value.length - 1];
    input.value = '';
    error.value = '';
    sending.value = true;
    try {
        await chatModel(
            { ...target.value, model: model.value },
            [...history, { role: 'user', content }],
            currentNode.value,
            request.signal,
            (text, reasoning) => {
                reply.content += text;
                reply.reasoning += reasoning;
            },
        );
        if (!reply.content.trim()) throw new Error(i18n.global.t('aiTools.chat.failed'));
        user.complete = true;
        reply.complete = true;
    } catch (e) {
        if (controller !== request) return;
        error.value = request.signal.aborted
            ? i18n.global.t('aiTools.chat.interrupted')
            : e instanceof Error
              ? e.message
              : i18n.global.t('aiTools.chat.failed');
        if (!reply.content && !reply.reasoning) messages.value.pop();
        input.value ||= content;
    } finally {
        if (controller === request) {
            sending.value = false;
            controller = undefined;
        }
    }
};
const onKeydown = (event: KeyboardEvent) => {
    if (event.key === 'Enter' && !event.shiftKey && !event.isComposing) {
        event.preventDefault();
        void send();
    }
};
watch([currentNode, () => route.fullPath], close);
onBeforeUnmount(close);
defineExpose({ open });
</script>

<style scoped lang="scss">
.model-chat {
    display: flex;
    flex-direction: column;
    height: calc(100dvh - 140px);
    min-height: 260px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 12px;
    overflow: hidden;
    background: var(--el-bg-color);
}
.chat-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 12px 16px;
    border-bottom: 1px solid var(--el-border-color-lighter);
}
.chat-model-select {
    position: relative;
    width: max-content;
    max-width: calc(100% - 32px);
    min-width: 0;
    font-size: 12px;
    .chat-model-width {
        display: block;
        padding: 0 28px 0 8px;
        line-height: 24px;
        overflow: hidden;
        white-space: nowrap;
        visibility: hidden;
    }
    :deep(.el-select) {
        position: absolute;
        inset: 0;
        width: 100%;
        min-width: 0;
    }
    :deep(.el-select__wrapper) {
        border-radius: 12px;
    }
    :deep(.el-select__caret) {
        font-size: 12px;
    }
}
.chat-mode {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
    padding: 6px 10px;
    border-radius: 6px;
    background: var(--el-fill-color-light);
    font-size: 13px;
    font-weight: 500;
}
.chat-messages {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    overscroll-behavior: contain;
    overflow-anchor: none;
}
.chat-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100%;
    padding: 24px;
    box-sizing: border-box;
    text-align: center;
    h3 {
        margin: 16px 0 8px;
        font-size: 17px;
        font-weight: 600;
    }
    p {
        margin: 0;
        color: var(--el-text-color-secondary);
        font-size: 13px;
        line-height: 1.6;
    }
}
.chat-empty-icon {
    display: grid;
    place-items: center;
    width: 48px;
    height: 48px;
    border-radius: 14px;
    background: var(--el-fill-color-light);
    color: var(--el-text-color-regular);
    font-size: 24px;
}
.chat-transcript,
.chat-bottom {
    width: 100%;
    max-width: 820px;
    margin: 0 auto;
    box-sizing: border-box;
    padding: 24px;
}
.chat-message {
    margin-bottom: 28px;
    overflow-wrap: anywhere;
}
.chat-role {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin-bottom: 8px;
}
.chat-message.user {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    margin-left: 48px;
    .chat-message-body {
        max-width: 100%;
        padding: 12px 16px;
        border-radius: 16px 16px 4px 16px;
        background: var(--el-fill-color-light);
    }
}
.chat-message.assistant :deep(.md-editor-preview-wrapper) {
    padding: 0;
}
.chat-message.assistant :deep(.md-editor) {
    background: transparent;
}
.chat-message.assistant :deep(.md-editor-preview) > :first-child {
    margin-top: 0;
}
.chat-text {
    white-space: pre-wrap;
    line-height: 1.7;
}
.chat-reasoning {
    color: var(--el-text-color-secondary);
    margin-bottom: 12px;
    border-left: 2px solid var(--el-border-color);
    padding-left: 12px;
    font-size: 13px;
    summary {
        cursor: pointer;
        padding: 4px 0;
    }
}
.chat-generating {
    display: inline-flex;
    align-items: center;
    gap: 8px;
}
.chat-generating-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--el-color-primary);
}
.chat-bottom {
    flex-shrink: 0;
    padding-top: 0;
}
.chat-notice {
    margin: 0 0 6px;
    border: 1px solid var(--el-color-warning-light-7);
    border-radius: 8px;
    :deep(.el-alert__title) {
        font-size: 12px;
        line-height: 1.5;
    }
}
.chat-bottom > .el-alert:not(.chat-notice) {
    margin-bottom: 12px;
}
.chat-composer {
    padding: 12px;
    border: 1px solid transparent;
    border-radius: 16px;
    background: var(--el-fill-color-light);
    transition: border-color 0.15s;
    &:focus-within {
        border-color: var(--el-color-primary-light-5);
    }
    :deep(.el-textarea__inner) {
        padding: 0 4px;
        background: transparent;
        box-shadow: none;
        resize: none;
        line-height: 1.7;
    }
}
.chat-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 10px;
    :deep(.el-button .el-icon) {
        font-size: 12px;
    }
}
.chat-composer-model {
    font-size: 12px;
    min-width: 0;
}
@media (max-width: 768px) {
    .model-chat {
        height: calc(100dvh - 125px);
    }
    .chat-toolbar {
        padding: 10px;
        gap: 8px;
    }
    .chat-transcript {
        padding: 16px;
    }
    .chat-bottom {
        padding: 0 12px 12px;
    }
    .chat-message.user {
        margin-left: 24px;
    }
}
</style>
