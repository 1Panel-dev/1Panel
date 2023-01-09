<template>
    <div>
        <slot></slot>
    </div>
    <div class="footer flx-justify-between">
        <div class="footer-left">
            <a href="https://fit2cloud.com/" target="_blank">杭州飞致云信息科技有限公司</a>
        </div>
        <div class="footer-right">
            <span>Version V1</span>
        </div>
    </div>
</template>

<style scoped lang="scss">
@import './index.scss';

.footer-left {
    margin-left: 20px;
}

.footer-right {
    float: right;
    margin-right: 20px;
}
</style>
