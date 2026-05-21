import { useGlobalStore } from '@/composables/useGlobalStore';
import { handleAuthResponseCode, handleAuthResponseStatus } from '@/utils/auth-response';

export const checkStreamAuth = async (url: string, currentNode?: string) => {
    const { currentNode: storeCurrentNode, language } = useGlobalStore();
    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), 5000);
    try {
        const res = await fetch(url.replace(/^ws/, 'http'), {
            credentials: 'include',
            headers: {
                'Accept-Language': language.value,
                CurrentNode: encodeURIComponent(currentNode || storeCurrentNode.value),
            },
            signal: controller.signal,
        });
        const statusResult = handleAuthResponseStatus(res.status);
        if (statusResult.handled) {
            return statusResult.message;
        }
        const contentType = res.headers.get('content-type') || '';
        if (!contentType.includes('application/json')) {
            await res.body?.cancel();
            return '';
        }
        const data = await res.json();
        const codeResult = handleAuthResponseCode(data);
        if (codeResult.handled) {
            return codeResult.message;
        }
        return '';
    } catch {
        return '';
    } finally {
        window.clearTimeout(timeout);
    }
};
