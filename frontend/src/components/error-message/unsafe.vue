<template>
    <div class="not-container">
        <img src="@/assets/images/unsafe.svg" class="not-img" alt="404" />
        <div class="not-detail">
            <h2>{{ $t('commons.login.notSafe') }}</h2>
            <h4>{{ $t('commons.login.safeEntrance1') }}</h4>
            <div>
                <h4>{{ $t('commons.login.safeEntrance2') }}</h4>
                <div style="cursor: pointer; float: left">
                    <el-icon color="#409EFC" style="margin-left: 5px; margin-top: 33px" :size="18" @click="onCopy()">
                        <DocumentCopy />
                    </el-icon>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts" name="404">
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import useClipboard from 'vue-clipboard3';
const { toClipboard } = useClipboard();

const onCopy = async () => {
    try {
        await toClipboard('1pctl user-info');
        MsgSuccess(i18n.global.t('commons.msg.copySuccess'));
    } catch (e) {
        MsgError(i18n.global.t('commons.msg.copyfailed'));
    }
};
</script>

<style scoped lang="scss">
.not-container {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    .not-img {
        margin-top: 300px;
    }
    .not-detail {
        margin-top: 300px;
        display: flex;
        flex-direction: column;
        h2,
        h4 {
            padding: 0;
            margin: 0;
        }
        h2 {
            font-size: 60px;
            color: #434e59;
        }
        h4 {
            margin: 30px 0 20px;
            float: left;
            font-size: 19px;
            font-weight: normal;
            color: #848587;
        }
    }
}
</style>
