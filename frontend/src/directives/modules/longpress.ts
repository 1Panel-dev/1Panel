/**
 * v-longpress
 * 长按指令，长按时触发事件
 */
import type { Directive, DirectiveBinding } from 'vue';

const directive: Directive = {
    mounted(el: HTMLElement, binding: DirectiveBinding) {
        if (typeof binding.value !== 'function') {
            throw 'callback must be a function';
        }
        // 定义变量
        let pressTimer: any = null;
        // 创建计时器（ 2秒后执行函数 ）
        const start = (e: any) => {
            if (e.button) {
                if (e.type === 'click' && e.button !== 0) {
                    return;
                }
            }
            if (pressTimer === null) {
                pressTimer = setTimeout(() => {
                    handler(e);
                }, 1000);
            }
        };
        // 取消计时器
        const cancel = () => {
            if (pressTimer !== null) {
                clearTimeout(pressTimer);
                pressTimer = null;
            }
        };
        // 运行函数
        const handler = (e: MouseEvent | TouchEvent) => {
            binding.value(e);
        };
        // 添加事件监听器
        el.addEventListener('mousedown', start);
        el.addEventListener('touchstart', start);
        // 取消计时器
        el.addEventListener('click', cancel);
        el.addEventListener('mouseout', cancel);
        el.addEventListener('touchend', cancel);
        el.addEventListener('touchcancel', cancel);
    },
};

export default directive;
