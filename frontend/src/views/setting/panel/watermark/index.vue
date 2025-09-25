<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.watermark')" @close="handleClose" size="large">
        <el-watermark
            class="watermark"
            :content="loadContent()"
            :font="{
                fontSize: form.fontSize,
                color: form.color,
            }"
            :rotate="form.rotate"
            :gap="[form.gap, form.gap]"
        >
            <el-form
                ref="formRef"
                label-position="top"
                :model="form"
                @submit.prevent
                v-loading="loading"
                :rules="rules"
            >
                <el-form-item :label="$t('setting.watermarkContent')" prop="content">
                    <el-input clearable v-model.trim="form.content" />
                    <span class="input-help">{{ $t('setting.contentHelper', ['${nodeName} ${nodeAddr}']) }}</span>
                </el-form-item>
                <el-form-item :label="$t('setting.watermarkColor')" prop="color">
                    <el-color-picker v-model="form.color" show-alpha />
                </el-form-item>
                <el-form-item :label="$t('setting.watermarkFont')" prop="fontSize">
                    <el-slider v-model="form.fontSize" :min="12" :max="100" />
                </el-form-item>
                <el-form-item :label="$t('setting.watermarkRotate')" prop="rotate">
                    <el-slider v-model="form.rotate" :min="-180" :max="180" />
                </el-form-item>
                <el-form-item :label="$t('setting.watermarkGap')" prop="gap">
                    <el-input-number class="number-input" v-model.number="form.gap" />
                </el-form-item>
            </el-form>
        </el-watermark>
        <template #footer>
            <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button @click="setDefault">{{ $t('website.setDefault') }}</el-button>
            <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox, FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { GlobalStore } from '@/store';
import { updateXpackSettingByKey } from '@/utils/xpack';
const globalStore = GlobalStore();

const emit = defineEmits<{ (e: 'search'): void }>();

const drawerVisible = ref();
const loading = ref();

let form = reactive({
    color: 'rgba(0, 0, 0, 0.15)',
    fontSize: 16,
    content: '${nodeName} - ${nodeAddr}',
    rotate: -22,
    gap: 100,
});

const rules = reactive({
    content: [Rules.requiredInput],
});
const formRef = ref<FormInstance>();

const acceptParams = (watermark: string): void => {
    if (watermark) {
        const parsedData = JSON.parse(watermark);
        form = reactive(parsedData);
    }
    drawerVisible.value = true;
};

const loadContent = () => {
    let itemName = form.content.replaceAll(
        '${nodeName}',
        globalStore.currentNode === 'local' ? globalStore.getMasterAlias() : globalStore.currentNode,
    );
    itemName = itemName.replaceAll('${nodeAddr}', globalStore.currentNodeAddr);
    return itemName;
};

const setDefault = () => {
    form.color = 'rgba(0, 0, 0, 0.15)';
    form.fontSize = 16;
    form.content = '${nodeName} - ${nodeAddr}';
    form.rotate = -22;
    form.gap = 100;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('setting.watermarkOpenHelper'), i18n.global.t('setting.watermark'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        }).then(async () => {
            loading.value = true;
            let itemVal = JSON.stringify(form);
            await updateXpackSettingByKey('Watermark', itemVal)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    globalStore.watermark = {
                        color: form.color,
                        fontSize: form.fontSize,
                        content: form.content,
                        rotate: form.rotate,
                        gap: form.gap,
                    };
                    handleClose();
                    return;
                })
                .catch(() => {
                    loading.value = false;
                });
        });
    });
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.number-input {
    width: 100%;
}
.watermark {
    height: calc(100vh - 160px);
}
</style>
