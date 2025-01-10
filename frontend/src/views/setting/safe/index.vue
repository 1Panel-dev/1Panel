<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.safe')" :divider="true">
            <template #main>
                <el-form
                    :model="form"
                    v-loading="loading"
                    :label-position="mobile ? 'top' : 'left'"
                    label-width="auto"
                    class="sm:w-full md:w-4/5 lg:w-3/5 2xl:w-1/2 max-w-max ml-8"
                >
                    <el-form-item :label="$t('setting.panelPort')" prop="serverPort">
                        <el-input disabled v-model.number="form.serverPort">
                            <template #append>
                                <el-button @click="onChangePort" icon="Setting">
                                    {{ $t('commons.button.set') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('setting.bindInfo')" prop="bindAddress">
                        <el-input disabled v-model="form.bindAddress">
                            <template #append>
                                <el-button @click="onChangeBind" icon="Setting">
                                    {{ $t('commons.button.set') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('setting.entrance')">
                        <el-input type="password" disabled v-if="form.securityEntrance" v-model="form.securityEntrance">
                            <template #append>
                                <el-button @click="onChangeEntrance" icon="Setting">
                                    {{ $t('commons.button.set') }}
                                </el-button>
                            </template>
                        </el-input>
                        <el-input disabled v-if="!form.securityEntrance" v-model="unset">
                            <template #append>
                                <el-button @click="onChangeEntrance" icon="Setting">
                                    {{ $t('commons.button.set') }}
                                </el-button>
                            </template>
                        </el-input>
                        <span class="input-help">{{ $t('setting.entranceHelper') }}</span>
                    </el-form-item>

                    <el-form-item :label="$t('setting.noAuthSetting')">
                        <el-input disabled v-model="form.noAuthSetting">
                            <template #append>
                                <el-button @click="onChangeResponse" icon="Setting">
                                    {{ $t('commons.button.set') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>

                    <el-form-item :label="$t('setting.allowIPs')">
                        <div style="width: 100%" v-if="form.allowIPs">
                            <el-input
                                type="textarea"
                                :rows="3"
                                disabled
                                v-model="form.allowIPs"
                                style="width: calc(100% - 80px)"
                            />
                            <el-button class="append-button" @click="onChangeAllowIPs" icon="Setting">
                                {{ $t('commons.button.set') }}
                            </el-button>
                        </div>
                        <el-input disabled v-if="!form.allowIPs" v-model="unset">
                            <template #append>
                                <el-button @click="onChangeAllowIPs" icon="Setting">
                                    {{ $t('commons.button.set') }}
                                </el-button>
                            </template>
                        </el-input>
                        <span class="input-help">{{ $t('setting.allowIPsHelper') }}</span>
                    </el-form-item>

                    <el-form-item :label="$t('setting.bindDomain')">
                        <el-input disabled v-if="form.bindDomain" v-model="form.bindDomain">
                            <template #append>
                                <el-button @click="onChangeBindDomain" icon="Setting">
                                    {{ $t('commons.button.set') }}
                                </el-button>
                            </template>
                        </el-input>
                        <el-input disabled v-if="!form.bindDomain" v-model="unset">
                            <template #append>
                                <el-button @click="onChangeBindDomain" icon="Setting">
                                    {{ $t('commons.button.set') }}
                                </el-button>
                            </template>
                        </el-input>
                        <span class="input-help">{{ $t('setting.bindDomainHelper') }}</span>
                    </el-form-item>

                    <el-form-item :label="$t('setting.panelSSL')" prop="ssl">
                        <el-switch
                            @change="handleSSL"
                            v-model="form.ssl"
                            active-value="enable"
                            inactive-value="disable"
                        />
                        <span class="input-help">{{ $t('setting.https') }}</span>
                        <div v-if="form.ssl === 'enable' && sslInfo">
                            <el-tag>{{ $t('setting.domainOrIP') }} {{ sslInfo.domain }}</el-tag>
                            <el-tag style="margin-left: 5px">
                                {{ $t('setting.timeOut') }} {{ dateFormat('', '', sslInfo.timeout) }}
                            </el-tag>
                            <div>
                                <el-button link type="primary" @click="handleSSL">
                                    {{ $t('commons.button.view') }}
                                </el-button>
                            </div>
                        </div>
                    </el-form-item>

                    <el-form-item :label="$t('setting.expirationTime')" prop="expirationTime">
                        <el-input disabled v-model="form.expirationTime">
                            <template #append>
                                <el-button @click="onChangeExpirationTime" icon="Setting">
                                    {{ $t('commons.button.set') }}
                                </el-button>
                            </template>
                        </el-input>
                        <div>
                            <span class="input-help" v-if="form.expirationTime !== $t('setting.unSetting')">
                                {{ $t('setting.timeoutHelper', [loadTimeOut()]) }}
                            </span>
                            <span class="input-help" v-else>
                                {{ $t('setting.noneSetting') }}
                            </span>
                        </div>
                    </el-form-item>
                    <el-form-item :label="$t('setting.complexity')" prop="complexityVerification">
                        <el-switch
                            @change="onSaveComplexity"
                            v-model="form.complexityVerification"
                            active-value="enable"
                            inactive-value="disable"
                        />
                        <span class="input-help">
                            {{ $t('setting.complexityHelper') }}
                        </span>
                    </el-form-item>

                    <el-form-item :label="$t('setting.mfa')">
                        <el-switch
                            @change="handleMFA"
                            v-model="form.mfaStatus"
                            active-value="enable"
                            inactive-value="disable"
                        />
                        <span class="input-help">
                            {{ $t('setting.mfaHelper') }}
                        </span>
                    </el-form-item>
                </el-form>
            </template>
        </LayoutContent>

        <PortSetting ref="portRef" />
        <BindSetting ref="bindRef" />
        <MfaSetting ref="mfaRef" @search="search" />
        <SSLSetting ref="sslRef" @search="search" />
        <EntranceSetting ref="entranceRef" @search="search" />
        <TimeoutSetting ref="timeoutRef" @search="search" />
        <DomainSetting ref="domainRef" @search="search" />
        <AllowIPsSetting ref="allowIPsRef" @search="search" />
        <ResponseSetting ref="responseRef" @search="search()" />
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, computed } from 'vue';
import { ElForm, ElMessageBox } from 'element-plus';
import PortSetting from '@/views/setting/safe/port/index.vue';
import BindSetting from '@/views/setting/safe/bind/index.vue';
import ResponseSetting from '@/views/setting/safe/response/index.vue';
import SSLSetting from '@/views/setting/safe/ssl/index.vue';
import MfaSetting from '@/views/setting/safe/mfa/index.vue';
import TimeoutSetting from '@/views/setting/safe/timeout/index.vue';
import EntranceSetting from '@/views/setting/safe/entrance/index.vue';
import DomainSetting from '@/views/setting/safe/domain/index.vue';
import AllowIPsSetting from '@/views/setting/safe/allowips/index.vue';
import { dateFormat } from '@/utils/util';
import { updateSetting, getSettingInfo, getSystemAvailable, updateSSL, loadSSLInfo } from '@/api/modules/setting';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { Setting } from '@/api/interface/setting';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const loading = ref(false);
const entranceRef = ref();
const portRef = ref();
const bindRef = ref();
const timeoutRef = ref();
const mfaRef = ref();
const responseRef = ref();

const sslRef = ref();
const sslInfo = ref<Setting.SSLInfo>();
const domainRef = ref();
const allowIPsRef = ref();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const form = reactive({
    serverPort: 9999,
    ipv6: 'disable',
    bindAddress: '',
    ssl: 'disable',
    sslType: 'self',
    securityEntrance: '',
    expirationDays: 0,
    expirationTime: '',
    complexityVerification: 'disable',
    mfaStatus: 'disable',
    mfaInterval: 30,
    allowIPs: '',
    bindDomain: '',
    noAuthSetting: '200 - ' + i18n.global.t('setting.help200'),
    noAuthSettingValue: '200',
});

const unset = ref(i18n.global.t('setting.unSetting'));

const search = async () => {
    const res = await getSettingInfo();
    form.serverPort = Number(res.data.serverPort);
    form.ipv6 = res.data.ipv6;
    form.bindAddress = res.data.bindAddress;
    form.ssl = res.data.ssl;
    form.sslType = res.data.sslType;
    if (form.ssl === 'enable') {
        loadInfo();
    }
    form.securityEntrance = res.data.securityEntrance;
    form.expirationDays = Number(res.data.expirationDays);
    form.expirationTime = res.data.expirationTime;
    form.complexityVerification = res.data.complexityVerification;
    form.mfaStatus = res.data.mfaStatus;
    form.mfaInterval = Number(res.data.mfaInterval);
    form.allowIPs = res.data.allowIPs.replaceAll(',', '\n');
    form.bindDomain = res.data.bindDomain;
    form.noAuthSettingValue = res.data.noAuthSetting;
    if (res.data.noAuthSetting !== '200') {
        form.noAuthSetting = res.data.noAuthSetting + ' - ' + i18n.global.t('setting.error' + res.data.noAuthSetting);
    } else {
        form.noAuthSetting = res.data.noAuthSetting + ' - ' + i18n.global.t('setting.help200');
    }
};

const onSaveComplexity = async () => {
    let param = {
        key: 'ComplexityVerification',
        value: form.complexityVerification,
    };
    loading.value = true;
    await updateSetting(param)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const handleMFA = async () => {
    if (form.mfaStatus === 'enable') {
        mfaRef.value.acceptParams({ interval: form.mfaInterval });
        return;
    }
    ElMessageBox.confirm(i18n.global.t('setting.mfaClose'), i18n.global.t('setting.mfa'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        loading.value = true;
        await updateSetting({ key: 'MFAStatus', value: 'disable' })
            .then(() => {
                loading.value = false;
                search();
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onChangeEntrance = () => {
    entranceRef.value.acceptParams({ securityEntrance: form.securityEntrance });
};
const onChangePort = () => {
    portRef.value.acceptParams({ serverPort: form.serverPort });
};
const onChangeBind = () => {
    bindRef.value.acceptParams({ ipv6: form.ipv6, bindAddress: form.bindAddress });
};
const onChangeResponse = () => {
    responseRef.value.acceptParams({ noAuthSetting: form.noAuthSettingValue });
};
const onChangeBindDomain = () => {
    domainRef.value.acceptParams({ bindDomain: form.bindDomain });
};
const onChangeAllowIPs = () => {
    allowIPsRef.value.acceptParams({ allowIPs: form.allowIPs });
};
const handleSSL = async () => {
    if (form.ssl === 'enable') {
        let params = {
            ssl: form.ssl,
            sslType: form.sslType,
            sslInfo: sslInfo.value,
        };
        sslRef.value!.acceptParams(params);
        return;
    }
    ElMessageBox.confirm(i18n.global.t('setting.sslDisableHelper'), i18n.global.t('setting.sslDisable'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    })
        .then(async () => {
            await updateSSL({ ssl: 'disable', domain: '', sslType: form.sslType, key: '', cert: '', sslID: 0 });
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            let href = window.location.href;
            globalStore.isLogin = false;
            let address = href.split('://')[1];
            if (globalStore.entrance) {
                address = address.replaceAll('settings/safe', globalStore.entrance);
            } else {
                address = address.replaceAll('settings/safe', 'login');
            }
            window.location.href = `http://${address}`;
        })
        .catch(() => {
            form.ssl = 'enable';
        });
};

const loadInfo = async () => {
    await loadSSLInfo().then(async (res) => {
        sslInfo.value = res.data;
    });
};

const onChangeExpirationTime = async () => {
    timeoutRef.value.acceptParams({ expirationDays: form.expirationDays });
};

function loadTimeOut() {
    if (form.expirationDays === 0) {
        form.expirationTime = i18n.global.t('setting.unSetting');
        return i18n.global.t('setting.unSetting');
    }
    let staytimeGap = new Date(form.expirationTime).getTime() - new Date().getTime();
    if (staytimeGap < 0) {
        form.expirationTime = i18n.global.t('setting.unSetting');
        return i18n.global.t('setting.unSetting');
    }
    return Math.floor(staytimeGap / (3600 * 1000 * 24));
}

onMounted(() => {
    search();
    getSystemAvailable();
});
</script>

<style lang="scss" scoped>
.append-button {
    width: 80px;
    background-color: var(--el-fill-color-light);
    color: var(--el-color-info);
}
</style>
