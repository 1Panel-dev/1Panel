<template>
    <el-container>
        <el-aside>
            <Menu>
                <slot name="menu"></slot>
            </Menu>
        </el-aside>
        <el-container>
            <el-main>
                <div>
                    <View></View>
                </div>
            </el-main>
            <el-footer>
                <Footer>
                    <slot name="footer"></slot>
                </Footer>
            </el-footer>
        </el-container>
    </el-container>
</template>

<script setup lang="ts">
import Menu from './layout-menu.vue';
import Footer from './layout-footer.vue';
import View from './layout-view.vue';
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
