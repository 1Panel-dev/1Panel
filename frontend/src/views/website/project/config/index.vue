<template>
    <LayoutContent :header="'网站设置'" :back-name="'Website'">
        <el-tabs type="card">
            <el-tab-pane label="基本">
                <el-tabs tab-position="left" type="border-card">
                    <el-tab-pane label="域名设置">
                        <ComplexTable>
                            <template #toolbar>
                                <el-button type="primary" plain>{{ '新增域名' }}</el-button>
                            </template>
                            <el-table-column :label="'域名'" prop="backup"></el-table-column>
                            <el-table-column :label="'端口'" prop="remark"></el-table-column>
                        </ComplexTable>
                    </el-tab-pane>
                    <el-tab-pane label="目录设置">Config</el-tab-pane>
                    <el-tab-pane label="HTTPS">Role</el-tab-pane>
                    <el-tab-pane label="域名跳转">Task</el-tab-pane>
                    <el-tab-pane label="错误页面">Task</el-tab-pane>
                    <el-tab-pane label="过期时间">Task</el-tab-pane>
                    <el-tab-pane label="流量限制">Task</el-tab-pane>
                    <el-tab-pane label="默认文档">Task</el-tab-pane>
                </el-tabs>
            </el-tab-pane>
            <el-tab-pane label="优化">优化</el-tab-pane>
            <el-tab-pane label="反代">反代</el-tab-pane>
            <el-tab-pane label="安全">反代</el-tab-pane>
            <el-tab-pane label="备份">反代</el-tab-pane>
            <el-tab-pane label="日志">反代</el-tab-pane>
            <el-tab-pane label="源文">反代</el-tab-pane>
        </el-tabs>
    </LayoutContent>
</template>

<script setup lang="ts">
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
</script>
