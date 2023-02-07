<template>
    <el-card>
        <el-row :gutter="20">
            <el-col :lg="3" :xl="2">
                <div>
                    <el-tag effect="dark" type="success">{{ props.primaryDomain }}</el-tag>
                </div>
            </el-col>
            <el-col :lg="4" :xl="2">
                <div class="span-font">
                    <Status class="span-font" :key="props.status" :status="props.status"></Status>
                </div>
            </el-col>
            <el-col :lg="4" :xl="4">
                <div class="span-font">
                    <el-tag>
                        {{ $t('website.expireDate') }}:
                        <span v-if="isEver(props.expireDate)">
                            {{ $t('website.neverExpire') }}
                        </span>
                        <span v-else>{{ dateFormatSimple(props.expireDate) }}</span>
                    </el-tag>
                </div>
            </el-col>
        </el-row>
    </el-card>
</template>
<script lang="ts" setup>
import Status from '@/components/status/index.vue';
import { dateFormatSimple } from '@/utils/util';
const props = defineProps({
    primaryDomain: {
        type: String,
        default: '',
    },
    status: {
        type: String,
        default: '',
    },
    expireDate: {
        type: String,
        default: '',
    },
});

const isEver = (time: string) => {
    const expireDate = new Date(time);
    return expireDate < new Date('1970-01-02');
};
</script>

<style lang="scss">
.span-font {
    font-size: 14px;
}
</style>
