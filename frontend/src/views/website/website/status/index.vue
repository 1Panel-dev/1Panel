<template>
    <div class="app-status">
        <el-card>
            <div class="flex w-full flex-col gap-4 md:flex-row">
                <div class="flex flex-wrap gap-4">
                    <el-tag effect="dark" type="success">{{ props.primaryDomain }}</el-tag>
                    <span>
                        <Status class="span-font" :key="props.status" :status="props.status"></Status>
                    </span>
                    <span>
                        <el-tag type="info">
                            {{ $t('website.expireDate') }}:
                            <span v-if="isEver(props.expireDate)">
                                {{ $t('website.neverExpire') }}
                            </span>
                            <span v-else>{{ dateFormatSimple(props.expireDate) }}</span>
                        </el-tag>
                    </span>
                </div>
            </div>
        </el-card>
    </div>
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

<style lang="scss"></style>
