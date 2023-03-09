/*
	需求：实现一个drawer可拖拽指令。
	使用：在 Dom 上加上 v-draggable 即可
	<el-drawer v-drawerDrag />
*/
import type { Directive } from 'vue';
interface ElType extends HTMLElement {
    parentNode: any;
}
const drawerDrag: Directive = {
    mounted: function (el: ElType) {
        const minWidth = 400;
        const maxWidth = document.body.clientWidth;
        const dragDom = el.querySelector('.el-drawer');
        (dragDom as HTMLElement).style.overflow = 'auto';

        const resizeElL = document.createElement('div');
        dragDom.appendChild(resizeElL);
        resizeElL.style.cursor = 'w-resize';
        resizeElL.style.position = 'absolute';
        resizeElL.style.height = '100%';
        resizeElL.style.width = '10px';
        resizeElL.style.left = '0px';
        resizeElL.style.top = '0px';

        resizeElL.onmousedown = (e) => {
            const elW = dragDom.clientWidth;
            const EloffsetLeft = (dragDom as HTMLElement).offsetLeft;
            const clientX = e.clientX;
            document.onmousemove = function (e) {
                e.preventDefault();
                // 左侧鼠标拖拽位置
                if (clientX > EloffsetLeft && clientX < EloffsetLeft + 10) {
                    // 往左拖拽
                    if (clientX > e.clientX) {
                        if (dragDom.clientWidth < maxWidth) {
                            (dragDom as HTMLElement).style.width = elW + (clientX - e.clientX) + 'px';
                        } else {
                            (dragDom as HTMLElement).style.width = maxWidth + 'px';
                        }
                    }
                    // 往右拖拽
                    if (clientX < e.clientX) {
                        if (dragDom.clientWidth > minWidth) {
                            (dragDom as HTMLElement).style.width = elW - (e.clientX - clientX) + 'px';
                        } else {
                            (dragDom as HTMLElement).style.width = minWidth + 'px';
                        }
                    }
                }
            };
            // 拉伸结束
            document.onmouseup = function () {
                document.onmousemove = null;
                document.onmouseup = null;
            };
        };
    },
};
export default drawerDrag;
