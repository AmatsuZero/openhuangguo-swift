// 先运行 ../build.sh android，把 Go 编译出的 core 库放进 libs/ 和 src/main/jniLibs/。
plugins {
    id("com.android.library") version "8.13.2"
    id("org.jetbrains.kotlin.android") version "2.1.20"
}

android {
    namespace = "dev.openhuangguo"
    compileSdk = 36

    defaultConfig {
        minSdk = 21
        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
        consumerProguardFiles("consumer-rules.pro")
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    testOptions {
        unitTests.isReturnDefaultValues = true
    }
}

kotlin {
    compilerOptions {
        jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17)
    }
}

dependencies {
    implementation(files("libs/core.jar"))
    api("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.10.2")
    api("androidx.media3:media3-datasource:1.5.1")

    testImplementation("junit:junit:4.13.2")
    testImplementation("org.json:json:20250517")
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.10.2")

    androidTestImplementation("androidx.test:runner:1.6.2")
    androidTestImplementation("androidx.test.ext:junit:1.2.1")
}
