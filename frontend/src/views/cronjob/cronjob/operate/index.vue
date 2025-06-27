<template>
    <LayoutContent
        back-name="CronjobItem"
        :title="isCreate ? $t('cronjob.create') : $t('commons.button.edit') + ' - ' + form.name"
    >
        <template #main>
            <el-form ref="formRef" label-position="top" :model="form" :rules="rules">
                <el-row type="flex" justify="center" :gutter="20">
                    <el-col :span="20">
                        <el-card>
                            <el-form-item :label="$t('cronjob.taskType')" prop="type">
                                <el-select
                                    v-if="isCreate"
                                    class="mini-form-item"
                                    @change="changeType"
                                    v-model="form.type"
                                >
                                    <el-option value="shell" :label="$t('cronjob.shell')" />
                                    <el-option value="app" :label="$t('cronjob.app')" />
                                    <el-option value="website" :label="$t('cronjob.website')" />
                                    <el-option value="database" :label="$t('cronjob.database')" />
                                    <el-option value="directory" :label="$t('cronjob.directory')" />
                                    <el-option value="log" :label="$t('cronjob.log')" />
                                    <el-option value="curl" :label="$t('cronjob.curl')" />
                                    <el-option value="cutWebsiteLog" :label="$t('cronjob.cutWebsiteLog')" />
                                    <el-option value="clean" :label="$t('setting.diskClean')" />
                                    <el-option value="snapshot" :label="$t('cronjob.snapshot')" />
                                    <el-option value="ntp" :label="$t('cronjob.ntp')" />
                                </el-select>
                                <div v-else class="w-full">
                                    <el-tag>{{ $t('cronjob.' + form.type) }}</el-tag>
                                </div>
                                <span class="input-help logText" v-if="form.type === 'log'">
                                    {{ $t('cronjob.logHelper1') }}
                                    <el-link
                                        class="link"
                                        icon="Position"
                                        @click="goRouter('/logs/system')"
                                        type="primary"
                                    >
                                        {{ $t('firewall.quickJump') }}
                                    </el-link>
                                </span>
                                <span class="input-help logText" v-if="form.type === 'log'">
                                    {{ $t('cronjob.logHelper2') }}
                                    <el-link class="link" icon="Position" @click="goRouter('/logs/ssh')" type="primary">
                                        {{ $t('firewall.quickJump') }}
                                    </el-link>
                                </span>
                                <span class="input-help logText" v-if="form.type === 'log'">
                                    {{ $t('cronjob.logHelper3') }}
                                    <el-link
                                        class="link"
                                        icon="Position"
                                        @click="goRouter('/logs/website')"
                                        type="primary"
                                    >
                                        {{ $t('firewall.quickJump') }}
                                    </el-link>
                                </span>
                                <span class="input-help logText" v-if="form.type === 'ntp'">
                                    {{ $t('cronjob.ntp_helper') }}
                                    <el-link
                                        class="link"
                                        icon="Position"
                                        @click="goRouter('/toolbox/device')"
                                        type="primary"
                                    >
                                        {{ $t('firewall.quickJump') }}
                                    </el-link>
                                </span>
                            </el-form-item>
                            <el-form-item :label="$t('cronjob.taskName')" prop="name">
                                <el-input
                                    class="mini-form-item"
                                    :disabled="!isCreate"
                                    clearable
                                    v-model.trim="form.name"
                                />
                            </el-form-item>
                        </el-card>

                        <el-card class="mt-5">
                            <el-form-item :label="$t('cronjob.cronSpec')" prop="specCustom">
                                <el-checkbox :label="$t('container.custom')" v-model="form.specCustom" />
                            </el-form-item>
                            <div v-if="!form.specCustom">
                                <el-form-item prop="spec">
                                    <div v-for="(specObj, index) of form.specObjs" :key="index" style="width: 100%">
                                        <el-select
                                            class="specTypeClass"
                                            v-model="specObj.specType"
                                            @change="changeSpecType(index)"
                                        >
                                            <el-option
                                                v-for="item in specOptions"
                                                :key="item.label"
                                                :value="item.value"
                                                :label="item.label"
                                            />
                                        </el-select>
                                        <el-select
                                            v-if="specObj.specType === 'perWeek'"
                                            class="specClass"
                                            v-model="specObj.week"
                                        >
                                            <el-option
                                                v-for="item in weekOptions"
                                                :key="item.label"
                                                :value="item.value"
                                                :label="item.label"
                                            />
                                        </el-select>
                                        <el-input v-if="hasDay(specObj)" class="specClass" v-model.number="specObj.day">
                                            <template #append>
                                                <div class="append">{{ $t('cronjob.day') }}</div>
                                            </template>
                                        </el-input>
                                        <el-input
                                            v-if="hasHour(specObj)"
                                            class="specClass"
                                            v-model.number="specObj.hour"
                                        >
                                            <template #append>
                                                <div class="append">{{ $t('commons.units.hour') }}</div>
                                            </template>
                                        </el-input>
                                        <el-input
                                            v-if="specObj.specType !== 'perNSecond'"
                                            class="specClass"
                                            v-model.number="specObj.minute"
                                        >
                                            <template #append>
                                                <div class="append">{{ $t('commons.units.minute') }}</div>
                                            </template>
                                        </el-input>
                                        <el-input
                                            v-if="specObj.specType === 'perNSecond'"
                                            class="specClass"
                                            v-model.number="specObj.second"
                                        >
                                            <template #append>
                                                <div class="append">{{ $t('commons.units.second') }}</div>
                                            </template>
                                        </el-input>
                                        <el-popover
                                            placement="top-start"
                                            :title="$t('cronjob.nextTime')"
                                            width="200"
                                            trigger="click"
                                        >
                                            <div v-for="(time, index_t) of nextTimes" :key="index_t">
                                                <el-tag class="mt-2">{{ time }}</el-tag>
                                            </div>
                                            <template #reference>
                                                <el-button class="ml-5" @click="loadNext(specObj)" link type="primary">
                                                    {{ $t('commons.button.preview') }}
                                                </el-button>
                                            </template>
                                        </el-popover>
                                        <el-button
                                            class="ml-2.5"
                                            link
                                            type="primary"
                                            @click="handleSpecDelete(index)"
                                            v-if="form.specObjs.length > 1"
                                        >
                                            {{ $t('commons.button.delete') }}
                                        </el-button>
                                        <el-divider v-if="form.specObjs.length > 1" class="divider" />
                                    </div>
                                </el-form-item>
                                <el-button class="mb-3" @click="handleSpecAdd()">
                                    {{ $t('commons.button.add') }}
                                </el-button>
                            </div>

                            <div v-if="form.specCustom">
                                <el-form-item prop="spec">
                                    <div v-for="(spec, index) of form.specs" :key="index" class="w-full">
                                        <el-input class="specCustom" v-model="form.specs[index]" />
                                        <el-popover
                                            placement="top-start"
                                            :title="$t('cronjob.nextTime')"
                                            width="200"
                                            trigger="click"
                                        >
                                            <div v-for="(time, index_t) of nextTimes" :key="index_t">
                                                <el-tag class="mt-2">{{ time }}</el-tag>
                                            </div>
                                            <template #reference>
                                                <el-button class="ml-2.5" @click="loadNext(spec)" link type="primary">
                                                    {{ $t('commons.button.preview') }}
                                                </el-button>
                                            </template>
                                        </el-popover>
                                        <el-button
                                            class="ml-2.5"
                                            link
                                            type="primary"
                                            @click="handleSpecCustomDelete(index)"
                                            v-if="form.specs.length > 1"
                                        >
                                            {{ $t('commons.button.delete') }}
                                        </el-button>
                                        <el-divider v-if="form.specs.length > 1" class="divider" />
                                    </div>
                                </el-form-item>
                                <el-button class="mb-3" @click="handleSpecCustomAdd()">
                                    {{ $t('commons.button.add') }}
                                </el-button>
                            </div>
                        </el-card>

                        <el-card class="mt-5">
                            <el-row :gutter="20">
                                <LayoutCol v-if="isWebsite()">
                                    <el-form-item
                                        :label="form.type === 'website' ? $t('cronjob.website') : $t('menu.website')"
                                        prop="websiteList"
                                    >
                                        <el-select
                                            v-model="form.websiteList"
                                            multiple
                                            @change="
                                                form.websiteList = form.websiteList.includes('all')
                                                    ? ['all']
                                                    : form.websiteList
                                            "
                                        >
                                            <el-option
                                                :disabled="websiteOptions.length === 0"
                                                :label="$t('commons.table.all')"
                                                value="all"
                                            />
                                            <el-option
                                                v-for="(item, index) in websiteOptions"
                                                :key="index"
                                                :value="item.id + ''"
                                                :label="item.primaryDomain"
                                            >
                                                <span>{{ item.primaryDomain }}</span>
                                                <el-tag class="tagClass">
                                                    {{ item.alias }}
                                                </el-tag>
                                            </el-option>
                                        </el-select>
                                        <span class="input-help" v-if="form.type === 'cutWebsiteLog'">
                                            {{ $t('cronjob.cutWebsiteLogHelper') }}
                                        </span>
                                    </el-form-item>
                                </LayoutCol>
                                <LayoutCol v-if="form.type === 'app'">
                                    <el-form-item :label="$t('cronjob.app')" prop="appIdList">
                                        <el-select
                                            clearable
                                            v-model="form.appIdList"
                                            multiple
                                            @change="
                                                form.appIdList = form.appIdList.includes('all')
                                                    ? ['all']
                                                    : form.appIdList
                                            "
                                        >
                                            <el-option
                                                :disabled="appOptions.length === 0"
                                                :label="$t('commons.table.all')"
                                                value="all"
                                            />
                                            <div v-for="item in appOptions" :key="item.id">
                                                <el-option :value="item.id + ''" :label="item.name">
                                                    <span>{{ item.name }}</span>
                                                    <el-tag class="tagClass">
                                                        {{ item.key }}
                                                    </el-tag>
                                                </el-option>
                                            </div>
                                        </el-select>
                                    </el-form-item>
                                </LayoutCol>
                                <LayoutCol v-if="form.type === 'database'">
                                    <el-form-item :label="$t('cronjob.database')">
                                        <el-select v-model="form.dbType" @change="loadDatabases">
                                            <el-option label="MySQL" value="mysql" />
                                            <el-option label="Mariadb" value="mariadb" />
                                            <el-option label="PostgreSQL" value="postgresql" />
                                        </el-select>
                                    </el-form-item>
                                </LayoutCol>
                                <LayoutCol v-if="form.type === 'database'">
                                    <el-form-item :label="$t('cronjob.database')" prop="dbNameList">
                                        <el-select
                                            clearable
                                            v-model="form.dbNameList"
                                            multiple
                                            @change="
                                                form.dbNameList = form.dbNameList.includes('all')
                                                    ? ['all']
                                                    : form.dbNameList
                                            "
                                        >
                                            <el-option
                                                :disabled="dbInfo.dbs.length === 0"
                                                :label="$t('commons.table.all')"
                                                value="all"
                                            />
                                            <el-option
                                                v-for="item in dbInfo.dbs"
                                                :key="item.id"
                                                :value="item.id + ''"
                                                :label="item.name"
                                            >
                                                <span>{{ item.name }}</span>
                                                <el-tag class="tagClass">
                                                    {{
                                                        item.from === 'local'
                                                            ? $t('commons.table.local')
                                                            : $t('database.remote')
                                                    }}
                                                </el-tag>
                                                <el-tag class="tagClass">
                                                    {{ item.database }}
                                                </el-tag>
                                            </el-option>
                                        </el-select>
                                    </el-form-item>
                                </LayoutCol>
                                <LayoutCol v-if="form.type === 'directory'">
                                    <el-form-item :label="$t('commons.button.backup')">
                                        <el-radio-group v-model="form.isDir" class="w-full">
                                            <el-radio :value="true">{{ $t('file.dir') }}</el-radio>
                                            <el-radio :value="false">{{ $t('menu.files') }}</el-radio>
                                        </el-radio-group>
                                    </el-form-item>
                                </LayoutCol>
                                <LayoutCol v-if="form.type === 'curl'">
                                    <el-form-item :label="$t('cronjob.url')" prop="url">
                                        <el-input clearable v-model.trim="form.url" />
                                    </el-form-item>
                                </LayoutCol>
                            </el-row>

                            <el-row :gutter="20">
                                <el-col :span="24" v-if="hasScript()">
                                    <el-form-item>
                                        <el-checkbox @change="loadUserOptions(false)" v-model="form.inContainer">
                                            {{ $t('cronjob.containerCheckBox') }}
                                        </el-checkbox>
                                    </el-form-item>

                                    <el-row :gutter="20" v-if="form.inContainer">
                                        <LayoutCol>
                                            <el-form-item :label="$t('cronjob.containerName')" prop="containerName">
                                                <el-select
                                                    @change="loadUserOptions(false)"
                                                    v-model="form.containerName"
                                                >
                                                    <el-option
                                                        v-for="item in containerOptions"
                                                        :key="item"
                                                        :value="item"
                                                        :label="item"
                                                    />
                                                </el-select>
                                            </el-form-item>
                                        </LayoutCol>
                                        <LayoutCol>
                                            <el-form-item
                                                :label="$t('container.command')"
                                                prop="command"
                                                :rules="Rules.requiredInput"
                                            >
                                                <el-checkbox border v-model="form.isCustom">
                                                    {{ $t('container.custom') }}
                                                </el-checkbox>
                                                <el-select
                                                    v-if="!form.isCustom"
                                                    style="width: calc(100% - 100px)"
                                                    filterable
                                                    clearable
                                                    v-model="form.command"
                                                >
                                                    <el-option value="ash" label="/bin/ash" />
                                                    <el-option value="bash" label="/bin/bash" />
                                                    <el-option value="sh" label="/bin/sh" />
                                                </el-select>
                                                <el-input
                                                    clearable
                                                    v-else
                                                    style="width: calc(100% - 100px)"
                                                    v-model="form.command"
                                                />
                                            </el-form-item>
                                        </LayoutCol>
                                    </el-row>
                                    <el-row :gutter="20">
                                        <LayoutCol>
                                            <el-form-item :label="$t('commons.table.user')" prop="user">
                                                <el-select filterable v-model="form.user">
                                                    <div v-for="item in userOptions" :key="item">
                                                        <el-option :value="item" :label="item" />
                                                    </div>
                                                </el-select>
                                            </el-form-item>
                                        </LayoutCol>
                                        <LayoutCol v-if="!form.inContainer">
                                            <el-form-item :label="$t('cronjob.executor')" prop="executor">
                                                <el-checkbox border v-model="form.isCustom">
                                                    {{ $t('container.custom') }}
                                                </el-checkbox>
                                                <el-select
                                                    v-if="!form.isCustom"
                                                    style="width: calc(100% - 100px)"
                                                    v-model="form.executor"
                                                >
                                                    <el-option value="bash" label="bash" />
                                                    <el-option value="python" label="python" />
                                                    <el-option value="sh" label="sh" />
                                                </el-select>
                                                <el-input
                                                    clearable
                                                    v-else
                                                    style="width: calc(100% - 100px)"
                                                    v-model="form.executor"
                                                />
                                            </el-form-item>
                                        </LayoutCol>
                                    </el-row>

                                    <el-form-item :label="$t('cronjob.shellContent')" prop="scriptMode">
                                        <el-radio-group @change="form.script = ''" v-model="form.scriptMode">
                                            <el-radio value="input">{{ $t('commons.button.edit') }}</el-radio>
                                            <el-radio value="library">{{ $t('cronjob.library.library') }}</el-radio>
                                            <el-radio value="select">{{ $t('container.pathSelect') }}</el-radio>
                                        </el-radio-group>
                                    </el-form-item>
                                    <el-form-item class="-mt-4" v-if="form.scriptMode === 'input'" prop="script">
                                        <CodemirrorPro
                                            v-model="form.script"
                                            placeholder="#Define or paste the content of your shell file here"
                                            mode="javascript"
                                            :heightDiff="400"
                                        />
                                    </el-form-item>
                                    <el-row :gutter="20" class="-mt-4">
                                        <LayoutCol>
                                            <el-form-item prop="scriptID" v-if="form.scriptMode === 'library'">
                                                <el-select filterable v-model="form.scriptID">
                                                    <el-option
                                                        v-for="item in scriptOptions"
                                                        :key="item.id"
                                                        :value="item.id"
                                                        :label="item.name"
                                                    />
                                                </el-select>
                                            </el-form-item>
                                            <el-form-item prop="script" v-if="form.scriptMode === 'select'">
                                                <el-input
                                                    :placeholder="$t('commons.example') + '/tmp/test.sh'"
                                                    v-model="form.script"
                                                >
                                                    <template #prepend>
                                                        <FileList @choose="loadScriptDir" :dir="false"></FileList>
                                                    </template>
                                                </el-input>
                                            </el-form-item>
                                        </LayoutCol>
                                    </el-row>
                                </el-col>

                                <LayoutCol v-if="isDir() && form.isDir">
                                    <el-form-item :label="$t('cronjob.backupContent')" prop="sourceDir">
                                        <el-input v-model="form.sourceDir">
                                            <template #prepend>
                                                <FileList @choose="loadDir" :dir="true" :path="form.sourceDir" />
                                            </template>
                                        </el-input>
                                    </el-form-item>
                                </LayoutCol>
                                <LayoutCol v-if="isDir() && !form.isDir">
                                    <el-form-item :label="$t('cronjob.backupContent')" prop="files">
                                        <el-input>
                                            <template #prepend>
                                                <FileList @choose="loadFile" :dir="false" />
                                            </template>
                                        </el-input>
                                        <div class="w-full">
                                            <ComplexTable :show-header="false" :data="form.files" v-if="form.files">
                                                <el-table-column prop="val" />
                                                <el-table-column width="60">
                                                    <template #default="scope">
                                                        <el-button
                                                            link
                                                            type="primary"
                                                            @click="handleFileDelete(scope.$index)"
                                                        >
                                                            {{ $t('commons.button.delete') }}
                                                        </el-button>
                                                    </template>
                                                </el-table-column>
                                            </ComplexTable>
                                        </div>
                                    </el-form-item>
                                </LayoutCol>
                            </el-row>

                            <el-row :gutter="20" v-if="form.type === 'snapshot'">
                                <LayoutCol>
                                    <el-form-item prop="withImage">
                                        <el-checkbox v-model="form.withImage" :label="$t('cronjob.withImage')" />
                                    </el-form-item>
                                </LayoutCol>
                            </el-row>
                            <el-row :gutter="20" v-if="form.type === 'snapshot'">
                                <LayoutCol>
                                    <el-form-item :label="$t('cronjob.ignoreApp')" prop="ignoreAppIDs">
                                        <el-select v-model="form.ignoreAppIDs" multiple cleanable>
                                            <div v-for="item in appOptions" :key="item.id">
                                                <el-option :value="item.id" :label="item.name">
                                                    <span>{{ item.name }}</span>
                                                    <el-tag class="tagClass">
                                                        {{ item.key }}
                                                    </el-tag>
                                                </el-option>
                                            </div>
                                        </el-select>
                                    </el-form-item>
                                </LayoutCol>
                            </el-row>

                            <el-row :gutter="20">
                                <LayoutCol v-if="isBackup()">
                                    <el-form-item :label="$t('setting.backupAccount')" prop="sourceAccountItems">
                                        <el-select multiple v-model="form.sourceAccountItems" @change="changeAccount">
                                            <div v-for="item in backupOptions" :key="item.id">
                                                <el-option
                                                    v-if="item.type !== $t('setting.LOCAL')"
                                                    :value="item.id"
                                                    :label="item.name"
                                                >
                                                    {{ item.name }}
                                                    <el-tag class="tagClass" type="primary">{{ item.type }}</el-tag>
                                                </el-option>
                                                <el-option v-else :value="item.id" :label="item.type" />
                                            </div>
                                        </el-select>
                                        <span class="input-help logText">
                                            {{ $t('cronjob.targetHelper') }}
                                            <el-link
                                                class="link"
                                                icon="Position"
                                                @click="goRouter('/settings/backupaccount')"
                                                type="primary"
                                            >
                                                {{ $t('firewall.quickJump') }}
                                            </el-link>
                                        </span>
                                    </el-form-item>
                                </LayoutCol>
                                <LayoutCol v-if="isBackup()">
                                    <el-form-item :label="$t('cronjob.default_download_path')" prop="downloadAccountID">
                                        <el-select v-model="form.downloadAccountID">
                                            <div v-for="item in accountOptions" :key="item.id">
                                                <el-option
                                                    v-if="item.type !== $t('setting.LOCAL')"
                                                    :value="item.id"
                                                    :label="item.name"
                                                >
                                                    {{ item.name }}
                                                    <el-tag class="tagClass" type="primary">{{ item.type }}</el-tag>
                                                </el-option>
                                                <el-option v-else :value="item.id" :label="item.type" />
                                            </div>
                                        </el-select>
                                    </el-form-item>
                                </LayoutCol>
                            </el-row>

                            <el-row :gutter="20">
                                <LayoutCol v-if="isBackup() && !isDatabase()">
                                    <el-form-item :label="$t('setting.compressPassword')" prop="secret">
                                        <el-input v-model="form.secret" />
                                    </el-form-item>
                                </LayoutCol>
                                <LayoutCol>
                                    <el-form-item :label="$t('cronjob.retainCopies')" prop="retainCopies">
                                        <el-input-number
                                            class="selectClass"
                                            :min="1"
                                            step-strictly
                                            :step="1"
                                            v-model.number="form.retainCopies"
                                        ></el-input-number>
                                        <span v-if="isBackup()" class="input-help">
                                            {{ $t('cronjob.retainCopiesHelper1') }}
                                        </span>
                                        <span v-else class="input-help">{{ $t('cronjob.retainCopiesHelper') }}</span>
                                    </el-form-item>
                                </LayoutCol>
                            </el-row>
                            <el-row :gutter="20">
                                <LayoutCol :span="20" v-if="hasExclusionRules()">
                                    <el-form-item :label="$t('cronjob.exclusionRules')" prop="exclusionRules">
                                        <el-input
                                            :placeholder="$t('cronjob.rulesHelper')"
                                            clearable
                                            v-model="form.exclusionRules"
                                        />
                                        <span class="input-help">{{ $t('cronjob.exclusionRulesHelper') }}</span>
                                    </el-form-item>
                                </LayoutCol>
                            </el-row>
                        </el-card>

                        <el-card class="mt-5">
                            <div v-if="!globalStore.isIntl">
                                <el-row :gutter="20">
                                    <LayoutCol>
                                        <el-form-item prop="hasAlert">
                                            <el-checkbox v-model="form.hasAlert" :label="$t('xpack.alert.isAlert')" />
                                            <span class="input-help">{{ $t('xpack.alert.cronJobHelper') }}</span>

                                            <span class="input-help logText" v-if="form.hasAlert && !isProductPro">
                                                {{ $t('xpack.alert.licenseHelper') }}
                                                <el-link class="link" @click="toUpload" type="primary">
                                                    {{ $t('license.levelUpPro') }}
                                                </el-link>
                                            </span>
                                        </el-form-item>
                                    </LayoutCol>
                                    <LayoutCol>
                                        <el-form-item
                                            prop="alertCount"
                                            v-if="form.hasAlert && isProductPro"
                                            :label="$t('xpack.alert.alertCount')"
                                        >
                                            <el-input-number
                                                class="selectClass"
                                                :min="1"
                                                step-strictly
                                                :step="1"
                                                v-model.number="form.alertCount"
                                            ></el-input-number>
                                            <span class="input-help">{{ $t('xpack.alert.alertCountHelper') }}</span>
                                        </el-form-item>
                                    </LayoutCol>
                                </el-row>
                            </div>

                            <el-row :gutter="20">
                                <LayoutCol>
                                    <el-form-item :label="$t('cronjob.timeout')" prop="timeoutItem">
                                        <el-input type="number" class="selectClass" v-model.number="form.timeoutItem">
                                            <template #append>
                                                <el-select v-model="form.timeoutUint" style="width: 80px">
                                                    <el-option :label="$t('commons.units.second')" value="s" />
                                                    <el-option :label="$t('commons.units.minute')" value="m" />
                                                    <el-option :label="$t('commons.units.hour')" value="h" />
                                                </el-select>
                                            </template>
                                        </el-input>
                                    </el-form-item>
                                </LayoutCol>
                                <LayoutCol>
                                    <el-form-item :label="$t('cronjob.retryTimes')" prop="retryTimes">
                                        <el-input-number
                                            class="selectClass"
                                            :min="0"
                                            step-strictly
                                            :step="1"
                                            v-model.number="form.retryTimes"
                                        ></el-input-number>
                                        <span class="input-help">{{ $t('cronjob.retryTimesHelper') }}</span>
                                    </el-form-item>
                                </LayoutCol>
                            </el-row>
                        </el-card>

                        <el-form-item class="mt-5">
                            <el-button :disabled="loading" @click="goBack">
                                {{ $t('commons.button.back') }}
                            </el-button>
                            <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                                {{ $t('commons.button.confirm') }}
                            </el-button>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
        </template>
    </LayoutContent>

    <LicenseImport ref="licenseRef" />
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import FileList from '@/components/file-list/index.vue';
import { listBackupOptions } from '@/api/modules/backup';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Cronjob } from '@/api/interface/cronjob';
import { addCronjob, editCronjob, loadCronjobInfo, loadNextHandle, loadScriptOptions } from '@/api/modules/cronjob';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import LayoutCol from '@/components/layout-col/form.vue';
import { listDbItems } from '@/api/modules/database';
import { getWebsiteOptions } from '@/api/modules/website';
import { MsgError, MsgSuccess } from '@/utils/message';
import { useRouter } from 'vue-router';
import { listContainer } from '@/api/modules/container';
import { Database } from '@/api/interface/database';
import { listAppInstalled } from '@/api/modules/app';
import {
    loadDefaultSpec,
    loadDefaultSpecCustom,
    specOptions,
    transObjToSpec,
    transSpecToObj,
    weekOptions,
} from '../helper';
import { loadUsers } from '@/api/modules/toolbox';
import { loadContainerUsers } from '@/api/modules/container';
import { storeToRefs } from 'pinia';
import { GlobalStore } from '@/store';
import LicenseImport from '@/components/license-import/index.vue';
import { transferTimeToSecond } from '@/utils/util';
const router = useRouter();

const globalStore = GlobalStore();
const licenseRef = ref();
const { isProductPro } = storeToRefs(globalStore);
const loading = ref();
const nextTimes = ref([]);

const isCreate = ref();
const form = reactive<Cronjob.CronjobInfo>({
    id: 0,
    name: '',
    type: 'shell',
    specCustom: false,
    spec: '',
    specs: [],
    specObjs: [{ specType: 'perMonth', week: 1, day: 3, hour: 1, minute: 30, second: 30 }],

    executor: '',
    isExecutorCustom: false,
    script: '',
    scriptMode: 'input',
    isCustom: false,
    command: '',
    inContainer: false,
    containerName: '',
    user: '',

    scriptID: null,
    appID: '',
    website: '',
    exclusionRules: '',
    dbType: 'mysql',
    dbName: '',
    url: '',
    isDir: true,
    files: [],
    sourceDir: '',
    snapshotRule: { withImage: false, ignoreAppIDs: [] },
    ignoreAppIDs: [],
    withImage: false,

    sourceAccounts: [],
    downloadAccount: '',
    sourceAccountIDs: '',
    downloadAccountID: 0,
    sourceAccountItems: [],

    websiteList: [],
    appIdList: [],
    dbNameList: [],

    retainCopies: 7,
    retryTimes: 3,
    timeout: 3600,
    timeoutItem: 3600,
    timeoutUint: 's',
    status: '',
    secret: '',
    hasAlert: false,
    alertCount: 0,
    alertTitle: '',
});

const search = async () => {
    if (!isCreate.value) {
        loading.value = true;
        await loadCronjobInfo(form.id)
            .then((res) => {
                loading.value = false;
                form.name = res.data.name;
                form.type = res.data.type;
                form.specCustom = res.data.specCustom;
                form.spec = res.data.spec;
                form.specs = res.data.specs;
                if (!form.specCustom && form.spec) {
                    let objs = [];
                    for (const item of res.data.spec.split(',')) {
                        objs.push(transSpecToObj(item));
                    }
                    form.specObjs = objs || [];
                }
                if (form.specCustom && form.spec) {
                    form.specs = form.spec.split(',') || [];
                }

                form.script = res.data.script;
                form.scriptMode = res.data.scriptMode;

                form.containerName = res.data.containerName;
                form.user = res.data.user;
                if (form.containerName.length !== 0) {
                    form.inContainer = true;
                    form.command = res.data.command || 'sh';
                    form.isCustom = form.command !== 'sh' && form.command !== 'bash' && form.command !== 'ash';
                } else {
                    form.executor = res.data.executor || 'bash';
                    form.isCustom =
                        form.executor !== 'sh' &&
                        form.executor !== 'bash' &&
                        form.executor !== 'python' &&
                        form.executor !== 'python3';
                }

                form.scriptID = res.data.scriptID;
                form.appID = res.data.appID;
                form.appIdList = res.data.appID.split(',') || [];
                form.website = res.data.website;
                form.websiteList = res.data.website.split(',') || [];
                form.exclusionRules = res.data.exclusionRules;
                form.dbType = res.data.dbType;
                form.dbName = res.data.dbName;
                form.dbNameList = res.data.dbName.split(',') || [];
                form.url = res.data.url;
                form.withImage = res.data.snapshotRule.withImage;
                form.ignoreAppIDs = res.data.snapshotRule.ignoreAppIDs;

                form.isDir = res.data.isDir;
                form.sourceDir = res.data.sourceDir;
                if (!form.isDir) {
                    let files = form.sourceDir?.split(',') || [];
                    for (const item of files) {
                        form.files.push({ val: item });
                    }
                }

                form.sourceAccountIDs = res.data.sourceAccountIDs;
                form.downloadAccountID = res.data.downloadAccountID;
                if (form.sourceAccountIDs) {
                    let list = [];
                    form.sourceAccountItems = [];
                    list = form.sourceAccountIDs.split(',');
                    for (const item of list) {
                        if (item) {
                            form.sourceAccountItems.push(Number(item));
                        }
                    }
                }

                form.retainCopies = res.data.retainCopies;
                form.retryTimes = res.data.retryTimes;
                form.timeout = res.data.timeout;
                form.timeoutItem = res.data.timeout || 3600;
                form.secret = res.data.secret;
                form.hasAlert = res.data.alertCount > 0;
                form.alertCount = res.data.alertCount || 3;
                form.alertTitle = res.data.alertTitle;
            })
            .catch(() => {
                loading.value = false;
            });
    }
    loadBackups();
    loadAppInstalls();
    loadUserOptions(true);
    loadWebsites();
    loadContainers();
    loadScripts();
    if (form.dbType) {
        loadDatabases(form.dbType);
    } else {
        loadDatabases('mysql');
    }
};

const goRouter = async (path: string) => {
    router.push({ path: path });
};

const containerOptions = ref([]);
const websiteOptions = ref([]);
const backupOptions = ref([]);
const accountOptions = ref([]);
const appOptions = ref([]);
const userOptions = ref([]);
const scriptOptions = ref([]);

const dbInfo = reactive({
    isExist: false,
    name: '',
    version: '',
    dbs: [] as Array<Database.DbItem>,
});

const verifyScript = (rule: any, value: any, callback: any) => {
    if (!form.script || form.script.length === 0) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
        return;
    }
    callback();
};

const verifySpec = (rule: any, value: any, callback: any) => {
    if (form.specCustom) {
        if (form.specs.length === 0) {
            callback(new Error(i18n.global.t('commons.rule.requiredInput')));
            return;
        }
        for (let i = 0; i < form.specs.length; i++) {
            if (form.specs[i]) {
                continue;
            }
            callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
            return;
        }
        callback();
        return;
    }
    if (form.specObjs.length === 0) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
    }
    for (let i = 0; i < form.specObjs.length; i++) {
        let item = form.specObjs[i];
        if (
            !Number.isInteger(item.day) ||
            !Number.isInteger(item.hour) ||
            !Number.isInteger(item.minute) ||
            !Number.isInteger(item.second) ||
            !Number.isInteger(item.week)
        ) {
            callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
            return;
        }
        switch (item.specType) {
            case 'perMonth':
                if (
                    item.day < 0 ||
                    item.day > 31 ||
                    item.hour < 0 ||
                    item.hour > 23 ||
                    item.minute < 0 ||
                    item.minute > 59
                ) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
            case 'perNDay':
                if (
                    item.day < 0 ||
                    item.day > 366 ||
                    item.hour < 0 ||
                    item.hour > 23 ||
                    item.minute < 0 ||
                    item.minute > 59
                ) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
            case 'perWeek':
                if (
                    item.week < 0 ||
                    item.week > 6 ||
                    item.hour < 0 ||
                    item.hour > 23 ||
                    item.minute < 0 ||
                    item.minute > 59
                ) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
            case 'perDay':
                if (item.hour < 0 || item.hour > 23 || item.minute < 0 || item.minute > 59) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
            case 'perNHour':
                if (item.hour < 0 || item.hour > 8784 || item.minute < 0 || item.minute > 59) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
            case 'perHour':
                if (item.minute < 0 || item.minute > 59) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
            case 'perNMinute':
                if (item.minute < 0 || item.minute > 527040) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
            case 'perNSecond':
                if (item.second < 0 || item.second > 31622400) {
                    callback(new Error(i18n.global.t('cronjob.cronSpecRule', [i + 1])));
                    return;
                }
                break;
        }
    }
    callback();
};

const verifyFiles = (rule: any, value: any, callback: any) => {
    if (!form.files || form.files.length === 0) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
        return;
    }
    callback();
};

const checkSendCount = (rule: any, value: any, callback: any) => {
    if (value === '') {
        callback();
    }
    const regex = /^(?:[1-9]|[12][0-9]|30)$/;
    if (!regex.test(value)) {
        return callback(new Error(i18n.global.t('commons.rule.numberRange', [1, 30])));
    }
    callback();
};

const rules = reactive({
    name: [Rules.requiredInput, Rules.noSpace],
    type: [Rules.requiredSelect],
    spec: [
        { validator: verifySpec, trigger: 'blur', required: true },
        { validator: verifySpec, trigger: 'change', required: true },
    ],

    script: [{ validator: verifyScript, trigger: 'blur', required: true }],
    scriptID: [Rules.requiredSelect],
    containerName: [Rules.requiredSelect],
    websiteList: [Rules.requiredSelect],
    appIdList: [Rules.requiredSelect],
    dbNameList: [Rules.requiredSelect],
    url: [Rules.requiredInput],
    files: [{ validator: verifyFiles, trigger: 'blur', required: true }],
    sourceDir: [Rules.requiredInput],
    sourceAccountItems: [Rules.requiredSelect],
    downloadAccountID: [Rules.requiredSelect],
    retainCopies: [Rules.number],
    retryTimes: [Rules.number],
    timeoutItem: [Rules.number],
    timeoutUint: [Rules.requiredSelect],
    alertCount: [Rules.integerNumber, { validator: checkSendCount, trigger: 'blur' }],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const loadDir = async (path: string) => {
    form.sourceDir = path;
};

const loadScriptDir = async (path: string) => {
    form.script = path;
};

const goBack = () => {
    router.push({ name: 'CronjobItem' });
};

const loadFile = async (path: string) => {
    for (const item of form.files) {
        if (item.val === path) {
            return;
        }
    }
    form.files.push({ val: path });
};

const hasDay = (item: any) => {
    return item.specType === 'perMonth' || item.specType === 'perNDay';
};
const hasHour = (item: any) => {
    return item.specType !== 'perHour' && item.specType !== 'perNMinute' && item.specType !== 'perNSecond';
};
const isWebsite = () => {
    return form.type === 'website' || form.type === 'cutWebsiteLog';
};
const isDir = () => {
    return form.type === 'directory';
};
const isDatabase = () => {
    return form.type === 'database';
};

const loadNext = async (spec: any) => {
    nextTimes.value = [];
    let specItem = '';
    if (!form.specCustom) {
        specItem = transObjToSpec(spec.specType, spec.week, spec.day, spec.hour, spec.minute, spec.second);
    } else {
        specItem = spec;
    }
    if (!specItem) {
        MsgError(i18n.global.t('cronjob.cronSpecRule2'));
        return;
    }
    const data = await loadNextHandle(specItem);
    nextTimes.value = data.data || [];
};

const loadScripts = async () => {
    const res = await loadScriptOptions();
    scriptOptions.value = res.data || [];
};

const loadDatabases = async (dbType: string) => {
    const data = await listDbItems(dbType);
    dbInfo.dbs = data.data || [];
};

const changeType = () => {
    form.specObjs = [loadDefaultSpec(form.type)];
    form.specs = [loadDefaultSpecCustom(form.type)];
};

const changeSpecType = (index: number) => {
    let item = form.specObjs[index];
    switch (item.specType) {
        case 'perMonth':
        case 'perNDay':
            item.day = 3;
            item.hour = 1;
            item.minute = 30;
            break;
        case 'perWeek':
            item.week = 1;
            item.hour = 1;
            item.minute = 30;
            break;
        case 'perDay':
        case 'perNHour':
            item.hour = 2;
            item.minute = 30;
            break;
        case 'perHour':
        case 'perNMinute':
            item.minute = 30;
            break;
        case 'perNSecond':
            item.second = 30;
            break;
    }
};

const handleSpecAdd = () => {
    let item = {
        specType: 'perWeek',
        week: 1,
        day: 0,
        hour: 1,
        minute: 30,
        second: 0,
    };
    form.specObjs.push(item);
};

const handleSpecCustomAdd = () => {
    form.specs.push('');
};

const handleSpecDelete = (index: number) => {
    form.specObjs.splice(index, 1);
};

const handleSpecCustomDelete = (index: number) => {
    form.specs.splice(index, 1);
};

const handleFileDelete = (index: number) => {
    form.files.splice(index, 1);
};

const loadBackups = async () => {
    const res = await listBackupOptions();
    let options = res.data || [];
    backupOptions.value = [];
    let local = 0;
    for (const item of options) {
        if (item.id === 0) {
            continue;
        }
        if (item.type == 'LOCAL') {
            local = item.id;
        }
        backupOptions.value.push({ id: item.id, type: i18n.global.t('setting.' + item.type), name: item.name });
    }
    if (!form.sourceAccountItems) {
        form.sourceAccountItems = local === 0 ? [local] : [];
    }
    changeAccount();
};

const changeAccount = async () => {
    accountOptions.value = [];
    let isInAccounts = false;
    for (const item of backupOptions.value) {
        let exist = false;
        for (const ac of form.sourceAccountItems) {
            if (item.id == ac) {
                exist = true;
                break;
            }
        }
        if (exist) {
            if (item.id === form.downloadAccountID) {
                isInAccounts = true;
            }
            accountOptions.value.push(item);
        }
    }
    if (!isInAccounts) {
        form.downloadAccountID = undefined;
    }
};

const loadUserOptions = async (isInit: boolean) => {
    if (!form.inContainer) {
        const res = await loadUsers();
        userOptions.value = res.data || [];
    } else {
        if (!isInit) {
            form.user = '';
        }
        if (form.containerName) {
            const res = await loadContainerUsers(form.containerName);
            userOptions.value = res.data || [];
        }
    }
};

const loadAppInstalls = async () => {
    const res = await listAppInstalled();
    appOptions.value = res.data || [];
};

const loadWebsites = async () => {
    const res = await getWebsiteOptions({});
    websiteOptions.value = res.data || [];
};

const loadContainers = async () => {
    const res = await listContainer();
    containerOptions.value = res.data || [];
};

function isBackup() {
    return (
        form.type === 'app' ||
        form.type === 'website' ||
        form.type === 'database' ||
        form.type === 'directory' ||
        form.type === 'snapshot' ||
        form.type === 'log'
    );
}

function hasExclusionRules() {
    return form.type === 'app' || form.type === 'website' || (form.type === 'directory' && form.isDir);
}

function hasScript() {
    return form.type === 'shell';
}

const onSubmit = async (formEl: FormInstance | undefined) => {
    let specs = [];
    if (!form.specCustom) {
        for (const item of form.specObjs) {
            const itemSpec = transObjToSpec(item.specType, item.week, item.day, item.hour, item.minute, item.second);
            if (itemSpec === '') {
                MsgError(i18n.global.t('cronjob.cronSpecHelper'));
                return;
            }
            specs.push(itemSpec);
        }
    } else {
        specs = form.specs;
    }
    if (!form.isDir) {
        let files = [];
        for (const item of form.files) {
            files.push(item.val);
        }
        form.sourceDir = files.join(',');
    }
    form.sourceAccountIDs = form.sourceAccountItems.join(',');
    form.spec = specs.join(',');
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!form.inContainer) {
            form.containerName = '';
        }
        form.timeout = transferTimeToSecond(form.timeoutItem + form.timeoutUint);
        if (form.appIdList) {
            form.appID = form.appIdList.join(',');
        }
        if (form.websiteList) {
            form.website = form.websiteList.join(',');
        }
        if (form.dbNameList) {
            form.dbName = form.dbNameList.join(',');
        }

        form.snapshotRule = { withImage: form.withImage, ignoreAppIDs: form.ignoreAppIDs };
        form.alertCount = form.hasAlert && isProductPro.value ? form.alertCount : 0;
        form.alertTitle =
            form.hasAlert && isProductPro.value
                ? i18n.global.t('cronjob.alertTitle', [i18n.global.t('cronjob.' + form.type), form.name])
                : '';
        if (!form) return;

        if (isCreate.value) {
            await addCronjob(form);
        } else {
            await editCronjob(form);
        }

        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        goBack();
    });
};

const toUpload = () => {
    licenseRef.value.acceptParams();
};

onMounted(() => {
    if (router.currentRoute.value.query.id) {
        isCreate.value = false;
        form.id = Number(router.currentRoute.value.query.id);
    } else {
        isCreate.value = true;
    }
    search();
});
</script>
<style scoped lang="scss">
.specClass {
    width: 17% !important;
    margin-left: 20px;
    .append {
        width: 20px;
    }
}
@media only screen and (max-width: 1000px) {
    .specClass {
        width: 100% !important;
        margin-top: 20px;
        margin-left: 0;
        .append {
            width: 43px;
        }
    }
}
.specTypeClass {
    width: 22% !important;
}
@media only screen and (max-width: 1000px) {
    .specTypeClass {
        width: 100% !important;
    }
}
.specCustom {
    width: 80%;
}
@media only screen and (max-width: 1000px) {
    .specCustom {
        width: 100% !important;
    }
}
.selectClass {
    width: 100%;
    padding-left: 0px;
}
.tagClass {
    float: right;
    margin-right: 10px;
    font-size: 12px;
    margin-top: 5px;
}
.logText {
    line-height: 22px;
    font-size: 12px;
    .link {
        font-size: 12px !important;
        margin-top: -3px;
    }
}

.divider {
    display: block;
    height: 1px;
    width: 100%;
    margin: 3px 0;
    border-top: 1px var(--el-border-color) var(--el-border-style);
}
</style>
