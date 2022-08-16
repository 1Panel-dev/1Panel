import { ref, onMounted, onUnmounted } from 'vue';

/**
 * @description 网络是否可用
 * */
export const useOnline = () => {
    const online = ref(true);
    const showStatus = (val: any) => {
        online.value = typeof val == 'boolean' ? val : val.target.online;
    };
    // 在页面加载后，设置正确的网络状态
    navigator.onLine ? showStatus(true) : showStatus(false);

    onMounted(() => {
        // 开始监听网络状态的变化
        window.addEventListener('online', showStatus);
        window.addEventListener('offline', showStatus);
    });

    onUnmounted(() => {
        // 移除监听网络状态的变化
        window.removeEventListener('online', showStatus);
        window.removeEventListener('offline', showStatus);
    });

    return { online };
};
