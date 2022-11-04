<template>
    <div>
        <span style="float: left">是否开启</span>
        <el-switch style="margin-left: 20px; float: left" v-model="form.slow_query_log" />
        <span style="margin-left: 30px; float: left">慢查询阈值</span>
        <div style="margin-left: 5px; float: left">
            <el-input v-model="form.long_query_time"></el-input>
        </div>
        <el-button style="margin-left: 20px">确认修改</el-button>
        <div v-if="form.slow_query_log === 'ON'">
            <codemirror
                :autofocus="true"
                placeholder="None data"
                :indent-with-tab="true"
                :tabSize="4"
                style="margin-top: 10px; max-height: 500px"
                :lineWrapping="true"
                :matchBrackets="true"
                theme="cobalt"
                :styleActiveLine="true"
                :extensions="extensions"
                v-model="slowLogs"
                :readOnly="true"
            />
        </div>
    </div>
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { reactive, ref } from 'vue';

const extensions = [javascript(), oneDark];
const slowLogs = ref();
const form = reactive({
    slow_query_log: 'OFF',
    long_query_time: 10,
});
</script>
