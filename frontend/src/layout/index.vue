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
            <el-footer v-if="themeConfig.footer">
                <Footer>
                    <slot name="footer"></slot>
                </Footer>
            </el-footer>
        </el-container>
    </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import Menu from './LayoutMenu.vue';
import Header from './LayoutHeader.vue';
import Footer from './LayoutFooter.vue';
import View from './LayoutView.vue';
import Content from './LayoutContent.vue';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();
const themeConfig = computed(() => globalStore.themeConfig);
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
