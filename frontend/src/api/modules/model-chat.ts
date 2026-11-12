import i18n from '@/lang';
import { getCookie } from '@/utils/auth';
import { handleAuthResponseCode, handleAuthResponseStatus } from '@/utils/auth-response';

export interface ModelChatTarget {
    kind: 'vllm' | 'account';
    id: number;
    name: string;
    model: string;
    models?: { id: string; name?: string }[];
}

export const supportsModelChat = (apiType: string) =>
    ['openai-completions', 'openai-responses', 'anthropic-messages', 'gemini-generate-content'].includes(apiType);

export interface ModelChatMessage {
    role: 'user' | 'assistant';
    content: string;
}

export const chatModel = async (
    target: ModelChatTarget,
    messages: ModelChatMessage[],
    node: string,
    signal: AbortSignal,
    onDelta: (content: string, reasoning: string) => void,
) => {
    const response = await fetch(
        `${import.meta.env.VITE_API_URL}${target.kind === 'vllm' ? '/xpack/vllm/chat' : '/ai/accounts/chat'}`,
        {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
                Accept: 'text/event-stream',
                'Accept-Language': i18n.global.locale.value,
                'X-CSRF-Token': getCookie('pcsrftoken') || '',
                CurrentNode: encodeURIComponent(node),
            },
            body: JSON.stringify({ id: target.id, model: target.model, messages }),
            signal,
        },
    );
    const auth = handleAuthResponseStatus(response.status);
    if (auth.handled) throw new Error(auth.message);
    if (response.headers.get('content-type')?.includes('application/json')) {
        const data = await response.json();
        const auth = handleAuthResponseCode(data);
        throw new Error(auth.message || data.message || i18n.global.t('aiTools.chat.failed'));
    }
    if (!response.ok || !response.body || !response.headers.get('content-type')?.includes('text/event-stream')) {
        await response.body?.cancel();
        throw new Error(`${i18n.global.t('aiTools.chat.failed')} (HTTP ${response.status})`);
    }
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    let data = '';
    const throwIncomplete = (reason = '') => {
        if (['max_output_tokens', 'MAX_TOKENS', 'length', 'max_tokens'].includes(reason)) {
            throw new Error(i18n.global.t('aiTools.chat.outputLimit'));
        }
        if (['content_filter', 'SAFETY', 'RECITATION', 'BLOCKLIST', 'PROHIBITED_CONTENT', 'SPII'].includes(reason)) {
            throw new Error(i18n.global.t('aiTools.chat.filtered'));
        }
        throw new Error(i18n.global.t('aiTools.chat.incomplete'));
    };
    try {
        while (true) {
            const chunk = await reader.read();
            if (chunk.done) throw new Error(i18n.global.t('aiTools.chat.interrupted'));
            buffer += decoder.decode(chunk.value, { stream: true });
            let end: number;
            while ((end = buffer.indexOf('\n')) >= 0) {
                const line = buffer.slice(0, end).replace(/\r$/, '');
                buffer = buffer.slice(end + 1);
                if (line.startsWith('data:')) {
                    data += line.slice(5).trimStart() + '\n';
                } else if (!line && data) {
                    const event = data.trim();
                    data = '';
                    if (event === '[DONE]') return;
                    const result = JSON.parse(event);
                    if (result.error) throw new Error(result.error.message || i18n.global.t('aiTools.chat.failed'));
                    switch (result.type) {
                        case 'response.output_text.delta':
                        case 'response.refusal.delta':
                            onDelta(result.delta || '', '');
                            break;
                        case 'response.reasoning_summary_text.delta':
                        case 'response.reasoning_text.delta':
                            onDelta('', result.delta || '');
                            break;
                        case 'content_block_start':
                            onDelta(result.content_block?.text || '', result.content_block?.thinking || '');
                            break;
                        case 'content_block_delta':
                            onDelta(result.delta?.text || '', result.delta?.thinking || '');
                            break;
                        case 'response.failed':
                            throw new Error(result.response?.error?.message || i18n.global.t('aiTools.chat.failed'));
                        case 'response.incomplete':
                            throwIncomplete(result.response?.incomplete_details?.reason);
                            break;
                        case 'message_delta':
                            if (
                                result.delta?.stop_reason &&
                                !['end_turn', 'stop_sequence', 'refusal'].includes(result.delta.stop_reason)
                            ) {
                                throwIncomplete(result.delta.stop_reason);
                            }
                            break;
                        case 'response.completed':
                        case 'message_stop':
                            return;
                        default: {
                            const candidate = result.candidates?.[0];
                            if (candidate) {
                                for (const part of candidate.content?.parts || []) {
                                    if (part.text)
                                        onDelta(part.thought ? '' : part.text, part.thought ? part.text : '');
                                }
                                if (candidate.finishReason) {
                                    if (candidate.finishReason !== 'STOP') throwIncomplete(candidate.finishReason);
                                    return;
                                }
                            } else {
                                if (result.promptFeedback?.blockReason)
                                    throwIncomplete(result.promptFeedback.blockReason);
                                const choice = result.choices?.[0];
                                const delta = choice?.delta;
                                onDelta(
                                    delta?.content || delta?.refusal || '',
                                    delta?.reasoning_content || delta?.reasoning || '',
                                );
                                if (choice?.finish_reason && choice.finish_reason !== 'stop')
                                    throwIncomplete(choice.finish_reason);
                            }
                        }
                    }
                }
            }
        }
    } finally {
        await reader.cancel().catch(() => {});
        reader.releaseLock();
    }
};
