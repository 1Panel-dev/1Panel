<template>
    <el-container>
        <el-aside>
            <Menu>
                <slot name="menu"></slot>
            </Menu>
        </el-aside>
        <el-container>
            <el-header>
                <Header>
                    <slot name="header"></slot>
                </Header>
            </el-header>
            <el-main>
                <Content>
                    <View></View>
                </Content>
            </el-main>
            <!-- <el-footer v-if="themeConfig.footer">
                <Footer>
                    <slot name="footer"></slot>
                </Footer>
            </el-footer> -->
        </el-container>
    </el-container>
</template>

<script setup lang="ts">
// import { computed } from 'vue';
import Menu from './layout-menu.vue';
import Header from './layout-header.vue';
// import Footer from './layout-footer.vue';
import View from './layout-view.vue';
import Content from './layout-content.vue';
// import { GlobalStore } from '@/store';
// const globalStore = GlobalStore();
// const themeConfig = computed(() => globalStore.themeConfig);
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
