<template>
    <div>
        <div>
            <el-row type="flex" justify="center">
                <h1>{{ loadErrInfo() }}</h1>
            </el-row>
        </div>
        <hr />
        <div>
            <el-row type="flex" justify="center"><span>nginx</span></el-row>
        </div>
    </div>
</template>

<script setup lang="ts">
const props = defineProps({
    code: String,
});
const loadErrInfo = () => {
    switch (props.code) {
        case '400':
            return '400 Bad Request';
        case '401':
            return '401 Unauthorized';
        case '403':
            return '403 Forbidden';
        case '404':
            return '404 Not Found';
        case '408':
            return '408 Request Timeout';
        case '416':
            return '416 Requested Not Satisfiable';
    }
};
</script>
