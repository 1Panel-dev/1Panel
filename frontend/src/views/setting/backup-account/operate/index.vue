<template>
    <DrawerPro v-model="drawerVisible" :header="title + $t('setting.backupAccount')" @close="handleClose" size="large">
        <el-form @submit.prevent ref="formRef" v-loading="loading" label-position="top" :model="dialogData.rowData">
            <el-form-item :label="$t('commons.table.name')" prop="name" :rules="Rules.requiredInput">
                <el-tag v-if="dialogData.title === 'edit'">{{ dialogData.rowData!.name }}</el-tag>
                <el-input v-else v-model="dialogData.rowData!.name" />
            </el-form-item>
            <el-form-item
                v-if="globalStore.isProductPro"
                :label="$t('setting.scope')"
                prop="isPublic"
                :rules="Rules.requiredSelect"
            >
                <el-tag v-if="dialogData.title === 'edit'">
                    {{ dialogData.rowData!.isPublic ? $t('setting.public') : $t('setting.private') }}
                </el-tag>
                <el-radio-group v-else v-model="dialogData.rowData!.isPublic">
                    <el-radio :value="true" size="large">{{ $t('setting.public') }}</el-radio>
                    <el-radio :value="false" size="large">{{ $t('setting.private') }}</el-radio>
                </el-radio-group>
                <span class="input-help">
                    {{ dialogData.rowData!.isPublic ? $t('setting.publicHelper') : $t('setting.privateHelper') }}
                </span>
            </el-form-item>
            <el-form-item :label="$t('commons.table.type')" prop="type" :rules="Rules.requiredSelect">
                <el-tag v-if="dialogData.title === 'edit'">{{ $t('setting.' + dialogData.rowData!.type) }}</el-tag>
                <el-select v-else v-model="dialogData.rowData!.type" @change="changeType">
                    <el-option :label="$t('setting.COS')" value="COS"></el-option>
                    <el-option :label="$t('setting.KODO')" value="KODO"></el-option>
                    <el-option :label="$t('setting.MINIO')" value="MINIO"></el-option>
                    <el-option :label="$t('setting.OneDrive')" value="OneDrive"></el-option>
                    <el-option :label="$t('setting.OSS')" value="OSS"></el-option>
                    <el-option :label="$t('setting.S3')" value="S3"></el-option>
                    <el-option :label="$t('setting.SFTP')" value="SFTP"></el-option>
                    <el-option :label="$t('setting.WebDAV')" value="WebDAV"></el-option>
                    <el-option :label="$t('setting.UPYUN')" value="UPYUN"></el-option>
                    <el-option :label="$t('setting.ALIYUN')" value="ALIYUN"></el-option>
                    <el-option :label="$t('setting.GoogleDrive')" value="GoogleDrive"></el-option>
                </el-select>
                <span v-if="isALIYUNYUN()" class="input-help">{{ $t('setting.ALIYUNHelper') }}</span>
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'S3'"
                :label="$t('setting.mode')"
                prop="varsJson.mode"
                :rules="Rules.requiredSelect"
            >
                <el-radio-group v-model="dialogData.rowData!.varsJson['mode']">
                    <el-radio value="virtual hosted">Virtual Hosted</el-radio>
                    <el-radio value="path">Path</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="hasAccessKey()" label="Access Key ID" prop="accessKey" :rules="Rules.requiredInput">
                <el-input clearable v-model.trim="dialogData.rowData!.accessKey" />
            </el-form-item>
            <el-form-item v-if="hasAccessKey()" label="Secret Key" prop="credential" :rules="Rules.requiredInput">
                <el-input show-password clearable v-model.trim="dialogData.rowData!.credential" />
            </el-form-item>
            <div v-if="isUPYUN()">
                <el-form-item :label="$t('setting.operator')" prop="accessKey" :rules="Rules.requiredInput">
                    <el-input clearable v-model.trim="dialogData.rowData!.accessKey" />
                </el-form-item>
                <el-form-item :label="$t('commons.login.password')" prop="credential" :rules="Rules.requiredInput">
                    <el-input show-password clearable v-model.trim="dialogData.rowData!.credential" />
                </el-form-item>
            </div>
            <el-form-item
                v-if="dialogData.rowData!.type === 'WebDAV'"
                :label="$t('setting.address')"
                prop="varsJson.address"
                :rules="Rules.requiredInput"
            >
                <el-input v-model="dialogData.rowData!.varsJson['address']" />
                <span class="input-help">
                    {{ $t('setting.WebDAVAlist') }}
                    <el-link
                        style="font-size: 12px; margin-left: 5px"
                        icon="Position"
                        @click="toWebDAVDoc()"
                        type="primary"
                    >
                        {{ $t('firewall.quickJump') }}
                    </el-link>
                </span>
            </el-form-item>
            <div v-if="dialogData.rowData!.type === 'SFTP'">
                <el-form-item :label="$t('setting.address')" prop="varsJson.address" :rules="Rules.host">
                    <el-input v-model.trim="dialogData.rowData!.varsJson['address']" clearable />
                </el-form-item>
                <el-form-item :label="$t('commons.table.port')" prop="varsJson.port" :rules="[Rules.port]">
                    <el-input-number :min="0" :max="65535" v-model.number="dialogData.rowData!.varsJson['port']" />
                </el-form-item>
            </div>
            <div v-if="hasPassword()">
                <el-form-item :label="$t('commons.login.username')" prop="accessKey" :rules="[Rules.requiredInput]">
                    <el-input v-model.trim="dialogData.rowData!.accessKey" />
                </el-form-item>

                <div v-if="dialogData.rowData!.type === 'SFTP'">
                    <el-form-item :label="$t('terminal.authMode')" prop="varsJson.authMode">
                        <el-radio-group v-model="dialogData.rowData!.varsJson['authMode']">
                            <el-radio value="password">{{ $t('terminal.passwordMode') }}</el-radio>
                            <el-radio value="key">{{ $t('terminal.keyMode') }}</el-radio>
                        </el-radio-group>
                    </el-form-item>
                </div>
                <div v-if="dialogData.rowData!.type === 'SFTP' && dialogData.rowData!.varsJson['authMode'] === 'key'">
                    <el-form-item :label="$t('terminal.key')" prop="credential" :rules="[Rules.requiredInput]">
                        <el-input type="textarea" v-model="dialogData.rowData!.credential" />
                    </el-form-item>
                    <el-form-item :label="$t('terminal.keyPassword')" prop="varsJson.passPhrase">
                        <el-input
                            type="password"
                            show-password
                            clearable
                            v-model="dialogData.rowData!.varsJson['passPhrase']"
                        />
                    </el-form-item>
                </div>
                <el-form-item
                    v-else
                    :label="$t('commons.login.password')"
                    prop="credential"
                    :rules="[Rules.requiredInput]"
                >
                    <el-input type="password" clearable show-password v-model.trim="dialogData.rowData!.credential" />
                </el-form-item>
            </div>
            <el-form-item v-if="hasRemember()" prop="rememberAuth">
                <el-checkbox v-model="dialogData.rowData!.rememberAuth">
                    {{ $t('terminal.rememberPassword') }}
                </el-checkbox>
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'COS'"
                label="Region"
                prop="varsJson.region"
                :rules="Rules.requiredInput"
            >
                <el-checkbox v-model="regionInput" :label="$t('container.input')" />
                <el-select v-if="!regionInput" v-model="dialogData.rowData!.varsJson['region']" filterable clearable>
                    <el-option v-for="item in cities" :key="item.value" :label="item.label" :value="item.value">
                        <span class="float-left">{{ item.label }}</span>
                        <span class="option-help">
                            {{ item.value }}
                        </span>
                    </el-option>
                </el-select>
                <el-input v-else v-model.trim="dialogData.rowData!.varsJson['region']" />
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'S3'"
                label="Region"
                prop="varsJson.region"
                :rules="Rules.requiredInput"
            >
                <el-input v-model.trim="dialogData.rowData!.varsJson['region']" />
            </el-form-item>
            <el-form-item
                v-if="hasAccessKey()"
                :label="dialogData.rowData!.type === 'KODO' ? $t('setting.domain') : 'Endpoint'"
                prop="varsJson.endpointItem"
                :rules="Rules.requiredInput"
            >
                <el-input v-model.trim="dialogData.rowData!.varsJson['endpointItem']">
                    <template #prepend>
                        <el-select v-model.trim="domainProto" class="p-w-100">
                            <el-option label="http" value="http" />
                            <el-option label="https" value="https" />
                        </el-select>
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item v-if="hasAccessKey()" label="Bucket" prop="bucket" :rules="Rules.requiredInput">
                <el-checkbox v-model="dialogData.rowData!.bucketInput" :label="$t('container.input')" />
                <el-input clearable v-if="dialogData.rowData!.bucketInput" v-model="dialogData.rowData!.bucket" />
                <div v-else class="w-full">
                    <el-select class="!w-4/5" v-model="dialogData.rowData!.bucket">
                        <el-option v-for="item in buckets" :key="item" :value="item" />
                    </el-select>
                    <el-button class="!w-1/5" plain @click="getBuckets(formRef)">
                        {{ $t('setting.loadBucket') }}
                    </el-button>
                </div>
            </el-form-item>
            <el-form-item
                v-if="isUPYUN()"
                :label="$t('setting.serviceName')"
                prop="bucket"
                :rules="Rules.requiredInput"
            >
                <el-input v-model="dialogData.rowData!.bucket" />
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'COS'"
                :label="$t('setting.scType')"
                prop="varsJson.scType"
                :rules="[Rules.requiredSelect]"
            >
                <el-select v-model="dialogData.rowData!.varsJson['scType']">
                    <el-option value="Standard" :label="$t('setting.scStandard')" />
                    <el-option value="Standard_IA" :label="$t('setting.scStandard_IA')" />
                    <el-option value="Archive" :label="$t('setting.scArchive')" />
                    <el-option value="Deep_Archive" :label="$t('setting.scDeep_Archive')" />
                </el-select>
                <el-alert
                    v-if="dialogData.rowData!.varsJson['scType'] === 'Archive' || dialogData.rowData!.varsJson['scType'] === 'Deep_Archive'"
                    class="mt-2.5"
                    :closable="false"
                    type="warning"
                    :title="$t('setting.archiveHelper')"
                />
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'OSS'"
                :label="$t('setting.scType')"
                prop="varsJson.scType"
                :rules="[Rules.requiredSelect]"
            >
                <el-select v-model="dialogData.rowData!.varsJson['scType']">
                    <el-option value="Standard" :label="$t('setting.scStandard')" />
                    <el-option value="IA" :label="$t('setting.scStandard_IA')" />
                    <el-option value="Archive" :label="$t('setting.scArchive')" />
                    <el-option value="ColdArchive" :label="$t('setting.scDeep_Archive')" />
                </el-select>
                <el-alert
                    v-if="dialogData.rowData!.varsJson['scType'] === 'Archive' || dialogData.rowData!.varsJson['scType'] === 'ColdArchive'"
                    class="mt-2.5"
                    :closable="false"
                    type="warning"
                    :title="$t('setting.archiveHelper')"
                />
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'S3'"
                :label="$t('setting.scType')"
                prop="varsJson.scType"
                :rules="[Rules.requiredSelect]"
            >
                <el-select v-model="dialogData.rowData!.varsJson['scType']">
                    <el-option value="STANDARD" :label="$t('setting.scStandard')" />
                    <el-option value="STANDARD_IA" :label="$t('setting.scStandard_IA')" />
                    <el-option value="GLACIER" :label="$t('setting.scArchive')" />
                    <el-option value="DEEP_ARCHIVE" :label="$t('setting.scDeep_Archive')" />
                </el-select>
                <el-alert
                    v-if="dialogData.rowData!.varsJson['scType'] === 'GLACIER' || dialogData.rowData!.varsJson['scType'] === 'DEEP_ARCHIVE'"
                    class="mt-2.5"
                    :closable="false"
                    type="warning"
                    :title="$t('setting.archiveHelper')"
                />
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'KODO'"
                :label="$t('cronjob.requestExpirationTime')"
                prop="varsJson.timeout"
            >
                <el-input-number
                    style="width: 200px"
                    :min="1"
                    step-strictly
                    :step="1"
                    v-model.number="dialogData.rowData!.varsJson['timeout']"
                ></el-input-number>
            </el-form-item>
            <div v-if="isALIYUNYUN()">
                <el-form-item label="Token" prop="varsJson.token">
                    <div class="!w-full">
                        <el-input
                            style="width: calc(100% - 80px)"
                            :rows="3"
                            type="textarea"
                            clearable
                            v-model.trim="dialogData.rowData!.varsJson['token']"
                        />
                        <el-button class="append-button" @click="loadFromTokenForAliyun()">
                            {{ $t('setting.analysis') }}
                        </el-button>
                        <span class="input-help">
                            {{ $t('setting.analysisHelper') }}
                            <el-link
                                style="font-size: 12px; margin-left: 5px"
                                icon="Position"
                                @click="toDoc(true)"
                                type="primary"
                            >
                                {{ $t('firewall.quickJump') }}
                            </el-link>
                        </span>
                    </div>
                </el-form-item>
                <el-form-item label="Drive ID" prop="varsJson.drive_id" :rules="Rules.requiredInput">
                    <el-input v-model.trim="dialogData.rowData!.varsJson['drive_id']" />
                </el-form-item>
                <el-form-item label="Refresh Token" prop="varsJson.refresh_token" :rules="Rules.requiredInput">
                    <el-input v-model.trim="dialogData.rowData!.varsJson['refresh_token']" />
                </el-form-item>
            </div>

            <div v-if="hasClient()">
                <el-form-item v-if="isOneDrive()">
                    <el-radio-group v-model="dialogData.rowData!.varsJson['isCN']" @change="changeClientFrom">
                        <el-radio-button :value="false">{{ $t('setting.isNotCN') }}</el-radio-button>
                        <el-radio-button :value="true">{{ $t('setting.isCN') }}</el-radio-button>
                    </el-radio-group>
                    <span class="input-help">
                        {{ $t('setting.onedrive_helper') }}
                        <el-link
                            style="font-size: 12px; margin-left: 5px"
                            icon="Position"
                            @click="toDoc(true)"
                            type="primary"
                        >
                            {{ $t('firewall.quickJump') }}
                        </el-link>
                    </span>
                </el-form-item>
                <el-form-item :label="$t('setting.client_id')" prop="varsJson.client_id" :rules="Rules.requiredInput">
                    <el-input v-model.trim="dialogData.rowData!.varsJson['client_id']" />
                </el-form-item>
                <el-form-item
                    :label="$t('setting.client_secret')"
                    prop="varsJson.client_secret"
                    :rules="Rules.requiredInput"
                >
                    <el-input v-model.trim="dialogData.rowData!.varsJson['client_secret']" />
                </el-form-item>
                <el-form-item
                    :label="$t('setting.redirect_uri')"
                    prop="varsJson.redirect_uri"
                    :rules="Rules.requiredInput"
                >
                    <el-input v-model.trim="dialogData.rowData!.varsJson['redirect_uri']" />
                </el-form-item>
                <el-form-item :label="$t('setting.code')" prop="varsJson.code" :rules="rules.driveCode">
                    <div class="!w-full">
                        <el-input
                            style="width: calc(100% - 80px)"
                            :rows="3"
                            type="textarea"
                            clearable
                            v-model.trim="dialogData.rowData!.varsJson['code']"
                        />
                        <el-button class="append-button" @click="jumpForCode(formRef)">
                            {{ $t('setting.loadCode') }}
                        </el-button>
                    </div>
                    <span class="input-help">
                        {{ $t('setting.codeHelper', [$t('setting.' + dialogData.rowData?.type)]) }}
                        <el-link
                            style="font-size: 12px; margin-left: 5px"
                            icon="Position"
                            @click="toDoc(false)"
                            type="primary"
                        >
                            {{ $t('firewall.quickJump') }}
                        </el-link>
                    </span>
                </el-form-item>
            </div>
            <el-form-item v-if="hasBackDir()" :label="$t('setting.backupDir')" prop="backupPath">
                <el-input clearable v-model.trim="dialogData.rowData!.backupPath" placeholder="/1panel" />
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'SFTP'"
                :label="$t('setting.backupDir')"
                prop="backupPath"
                :rules="[Rules.requiredInput]"
            >
                <el-input v-model.trim="dialogData.rowData!.backupPath" />
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'LOCAL'"
                :label="$t('setting.backupDir')"
                prop="backupPath"
                :rules="Rules.requiredInput"
            >
                <el-input v-model="dialogData.rowData!.backupPath">
                    <template #prepend>
                        <FileList @choose="loadDir" :dir="true"></FileList>
                    </template>
                </el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="loading" @click="handleClose">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Backup } from '@/api/interface/backup';
import { addBackup, editBackup, getClientInfo, listBucket } from '@/api/modules/backup';
import { cities } from './../helper';
import { deepCopy, spliceHttp, splitHttp } from '@/utils/util';
import { MsgSuccess } from '@/utils/message';
import { Base64 } from 'js-base64';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const loading = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const buckets = ref();
const clientInfo = ref();

const regionInput = ref();

const domainProto = ref('http');
const emit = defineEmits(['search']);
const rules = reactive({
    driveCode: [{ validator: checkDriveCode, required: true, trigger: 'blur' }],
});
function checkDriveCode(rule: any, value: any, callback: any) {
    if (!value) {
        return callback(new Error(i18n.global.t('setting.codeWarning')));
    }
    const reg = /^[A-Za-z0-9/_.-]+$/;
    if (!reg.test(value)) {
        return callback(new Error(i18n.global.t('setting.codeWarning')));
    }
    callback();
}

interface DialogProps {
    title: string;
    rowData?: Backup.BackupInfo;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    dialogData.value.rowData.varsJson = dialogData.value.rowData!.vars
        ? JSON.parse(dialogData.value.rowData!.vars)
        : {};
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    if (dialogData.value.title === 'create') {
        dialogData.value.rowData!.type = 'OSS';
        changeType();
        drawerVisible.value = true;
        return;
    }
    buckets.value = [];
    if (hasAccessKey()) {
        let itemJson = dialogData.value.rowData!.varsJson['endpoint'];
        if (dialogData.value.rowData!.type === 'KODO') {
            itemJson = dialogData.value.rowData!.varsJson['domain'];
        }
        let httpItem = splitHttp(itemJson);
        dialogData.value.rowData!.varsJson['endpointItem'] = httpItem.url;
        domainProto.value = httpItem.proto;
    }
    if (dialogData.value.rowData!.rememberAuth) {
        dialogData.value.rowData!.accessKey = Base64.decode(dialogData.value.rowData!.accessKey);
        dialogData.value.rowData!.credential = Base64.decode(dialogData.value.rowData!.credential);
    }
    if (dialogData.value.rowData!.varsJson['timeout'] === undefined) {
        dialogData.value.rowData!.varsJson['timeout'] = 1;
    }
    drawerVisible.value = true;
};
const toDoc = (isConf: boolean) => {
    let item = isConf ? '#32-onedrive' : '#33-onedrive';
    if (globalStore.isIntl) {
        item = isConf ? '#using-your-own-client-info-for-onedrive' : '#auth-code-of-onedrive';
    }
    window.open(globalStore.docsUrl + '/user_manual/settings/' + item, '_blank', 'noopener,noreferrer');
};
const toWebDAVDoc = () => {
    const uri = globalStore.isIntl ? '#webdav-with-alist' : '#34-webdav-alist';
    window.open(globalStore.docsUrl + '/user_manual/settings/' + uri, '_blank', 'noopener,noreferrer');
};
const jumpForCode = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    const result = await formEl.validateField('varsJson.client_id', callback);
    if (!result) {
        return;
    }
    const result1 = await formEl.validateField('varsJson.redirect_uri', callback);
    if (!result1) {
        return;
    }
    let client_id = dialogData.value.rowData.varsJson['client_id'];
    let redirect_uri = dialogData.value.rowData.varsJson['redirect_uri'];
    if (isOneDrive()) {
        let commonUrl = `response_type=code&client_id=${client_id}&redirect_uri=${redirect_uri}&scope=offline_access+Files.ReadWrite.All+User.Read`;
        if (!dialogData.value.rowData!.varsJson['isCN']) {
            window.open('https://login.microsoftonline.com/common/oauth2/v2.0/authorize?' + commonUrl, '_blank');
        } else {
            window.open('https://login.chinacloudapi.cn/common/oauth2/v2.0/authorize?' + commonUrl, '_blank');
        }
        return;
    }

    let url = `https://accounts.google.com/o/oauth2/auth/oauthchooseaccount?client_id=${client_id}&response_type=code&redirect_uri=${redirect_uri}&scope=openid%20profile%20https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fdrive%20https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fphotoslibrary&access_type=offline&prompt=consent&service=lso&o2v=1&ddm=1&flowName=GeneralOAuthFlow`;
    window.open(url, '_blank');
};
function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}

const loadFromTokenForAliyun = () => {
    const obj = JSON.parse(dialogData.value.rowData!.varsJson['token']);
    dialogData.value.rowData!.varsJson['drive_id'] = obj.default_drive_id;
    dialogData.value.rowData!.varsJson['refresh_token'] = obj.refresh_token;
};
const hasRemember = () => {
    return (
        dialogData.value.rowData!.type !== 'LOCAL' &&
        dialogData.value.rowData!.type !== 'OneDrive' &&
        dialogData.value.rowData!.type !== 'ALIYUN' &&
        dialogData.value.rowData!.type !== 'GoogleDrive'
    );
};
const hasClient = () => {
    let itemType = dialogData.value.rowData!.type;
    return itemType === 'OneDrive' || itemType === 'GoogleDrive';
};
const isOneDrive = () => {
    let itemType = dialogData.value.rowData!.type;
    return itemType === 'OneDrive';
};
const isUPYUN = () => {
    let itemType = dialogData.value.rowData!.type;
    return itemType === 'UPYUN';
};
const isALIYUNYUN = () => {
    let itemType = dialogData.value.rowData!.type;
    return itemType === 'ALIYUN';
};
const hasAccessKey = () => {
    let itemType = dialogData.value.rowData!.type;
    return itemType === 'COS' || itemType === 'KODO' || itemType === 'MINIO' || itemType === 'OSS' || itemType === 'S3';
};
const hasPassword = () => {
    let itemType = dialogData.value.rowData!.type;
    return itemType === 'SFTP' || itemType === 'WebDAV';
};

const hasBackDir = () => {
    let itemType = dialogData.value.rowData!.type;
    return itemType !== 'LOCAL' && itemType !== 'SFTP';
};

const loadDir = async (path: string) => {
    dialogData.value.rowData!.backupPath = path;
};

const changeType = async () => {
    buckets.value = [];
    dialogData.value.rowData!.varsJson = {};
    dialogData.value.rowData!.rememberAuth = false;
    switch (dialogData.value.rowData!.type) {
        case 'COS':
        case 'OSS':
        case 'S3':
            dialogData.value.rowData.varsJson['scType'] = 'Standard';
            dialogData.value.rowData.varsJson['mode'] = 'virtual hosted';
            break;
        case 'KODO':
            dialogData.value.rowData!.varsJson['timeout'] = 1;
            break;
        case 'OneDrive':
            dialogData.value.rowData.varsJson['isCN'] = false;
            const res = await getClientInfo('Onedrive');
            clientInfo.value = res.data;
            if (!dialogData.value.rowData.id) {
                dialogData.value.rowData.varsJson = {
                    isCN: false,
                    client_id: res.data.client_id,
                    client_secret: res.data.client_secret,
                    redirect_uri: res.data.redirect_uri,
                };
            }
            break;
        case 'GoogleDrive':
            const res2 = await getClientInfo('GoogleDrive');
            clientInfo.value = res2.data;
            if (!dialogData.value.rowData.id) {
                dialogData.value.rowData.varsJson = {
                    client_id: res2.data.client_id,
                    client_secret: res2.data.client_secret,
                    redirect_uri: res2.data.redirect_uri,
                };
            }
            break;
        case 'SFTP':
            dialogData.value.rowData.varsJson['port'] = 22;
            dialogData.value.rowData.varsJson['authMode'] = 'password';
            break;
    }
};
const changeClientFrom = () => {
    if (dialogData.value.rowData.varsJson['isCN']) {
        dialogData.value.rowData.varsJson = {
            isCN: true,
            client_id: '',
            client_secret: '',
            redirect_uri: '',
        };
    } else {
        dialogData.value.rowData.varsJson = {
            isCN: false,
            client_id: clientInfo.value.client_id,
            client_secret: clientInfo.value.client_secret,
            redirect_uri: clientInfo.value.redirect_uri,
        };
    }
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

const getBuckets = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    const result1 = await formEl.validateField('varsJson.endpointItem', callback);
    const result2 = await formEl.validateField('accessKey', callback);
    const result3 = await formEl.validateField('credential', callback);
    const result4 = await formEl.validateField('varsJson.region', callback);
    if (!result1 || !result2 || !result3 || !result4) {
        return;
    }
    loading.value = true;
    let item = deepCopy(dialogData.value.rowData!.varsJson);
    if (dialogData.value.rowData!.type === 'KODO') {
        item['domain'] = spliceHttp(domainProto.value, dialogData.value.rowData!.varsJson['endpointItem']);
    } else {
        item['endpoint'] = spliceHttp(domainProto.value, dialogData.value.rowData!.varsJson['endpointItem']);
    }
    item['endpointItem'] = undefined;
    listBucket({
        isPublic: dialogData.value.rowData!.isPublic,
        type: dialogData.value.rowData!.type,
        vars: JSON.stringify(item),
        accessKey: dialogData.value.rowData!.accessKey,
        credential: dialogData.value.rowData!.credential,
    })
        .then((res) => {
            loading.value = false;
            buckets.value = res.data;
        })
        .catch(() => {
            buckets.value = [];
            loading.value = false;
        });
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!dialogData.value.rowData) return;
        if (hasAccessKey()) {
            let itemEndpoint = spliceHttp(domainProto.value, dialogData.value.rowData!.varsJson['endpointItem']);
            if (dialogData.value.rowData!.type === 'KODO') {
                dialogData.value.rowData!.varsJson['domain'] = itemEndpoint;
            } else {
                dialogData.value.rowData!.varsJson['endpoint'] = itemEndpoint;
            }
            dialogData.value.rowData!.varsJson['endpointItem'] = itemEndpoint;
        }
        if (isALIYUNYUN()) {
            dialogData.value.rowData!.varsJson['token'] = undefined;
        }
        dialogData.value.rowData.vars = JSON.stringify(dialogData.value.rowData!.varsJson);
        loading.value = true;
        if (dialogData.value.title === 'create') {
            await addBackup(dialogData.value.rowData)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    drawerVisible.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
            return;
        }
        await editBackup(dialogData.value.rowData)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                drawerVisible.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.option-help {
    float: right;
    font-size: 12px;
    word-break: break-all;
    color: #8f959e;
}
</style>
