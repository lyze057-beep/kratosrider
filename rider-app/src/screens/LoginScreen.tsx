import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  Alert,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { authApi } from '../api/auth';
import { useAuthStore } from '../store/authStore';

type LoginMode = 'password' | 'code';

export const LoginScreen: React.FC = () => {
  const navigation = useNavigation();
  const { setAuth } = useAuthStore();
  
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [code, setCode] = useState('');
  const [loginMode, setLoginMode] = useState<LoginMode>('password');
  const [countdown, setCountdown] = useState(0);
  const [loading, setLoading] = useState(false);

  const handleSendCode = async () => {
    if (!phone || phone.length !== 11) {
      Alert.alert('提示', '请输入正确的手机号');
      return;
    }

    try {
      await authApi.sendCode({ phone, type: 'login' });
      Alert.alert('成功', '验证码已发送');
      setCountdown(60);
      const timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(timer);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    } catch (error: any) {
      Alert.alert('错误', error.response?.data?.message || '发送验证码失败');
    }
  };

  const handleLogin = async () => {
    if (!phone || phone.length !== 11) {
      Alert.alert('提示', '请输入正确的手机号');
      return;
    }

    if (loginMode === 'password' && !password) {
      Alert.alert('提示', '请输入密码');
      return;
    }

    if (loginMode === 'code' && !code) {
      Alert.alert('提示', '请输入验证码');
      return;
    }

    setLoading(true);
    try {
      let result;
      if (loginMode === 'password') {
        result = await authApi.loginByPassword({ phone, password });
      } else {
        result = await authApi.loginByPhone({ phone, code });
      }

      await setAuth(result.token, result.refresh_token, result.user_info);
      Alert.alert('成功', '登录成功');
    } catch (error: any) {
      Alert.alert('错误', error.response?.data?.message || '登录失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView contentContainerStyle={styles.scrollContent}>
        <View style={styles.header}>
          <Text style={styles.title}>美团骑手</Text>
          <Text style={styles.subtitle}>欢迎回来</Text>
        </View>

        <View style={styles.form}>
          <TextInput
            style={styles.input}
            placeholder="请输入手机号"
            keyboardType="phone-pad"
            maxLength={11}
            value={phone}
            onChangeText={setPhone}
          />

          {loginMode === 'password' ? (
            <TextInput
              style={styles.input}
              placeholder="请输入密码"
              secureTextEntry
              value={password}
              onChangeText={setPassword}
            />
          ) : (
            <View style={styles.codeRow}>
              <TextInput
                style={[styles.input, styles.codeInput]}
                placeholder="请输入验证码"
                keyboardType="number-pad"
                maxLength={6}
                value={code}
                onChangeText={setCode}
              />
              <TouchableOpacity
                style={[styles.codeBtn, countdown > 0 && styles.codeBtnDisabled]}
                onPress={handleSendCode}
                disabled={countdown > 0}
              >
                <Text style={styles.codeBtnText}>
                  {countdown > 0 ? `${countdown}s` : '获取验证码'}
                </Text>
              </TouchableOpacity>
            </View>
          )}

          <TouchableOpacity
            style={styles.switchMode}
            onPress={() => setLoginMode(loginMode === 'password' ? 'code' : 'password')}
          >
            <Text style={styles.switchModeText}>
              {loginMode === 'password' ? '验证码登录' : '密码登录'}
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.loginBtn, loading && styles.loginBtnDisabled]}
            onPress={handleLogin}
            disabled={loading}
          >
            <Text style={styles.loginBtnText}>
              {loading ? '登录中...' : '登录'}
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.registerBtn}
            onPress={() => navigation.navigate('Register' as never)}
          >
            <Text style={styles.registerBtnText}>还没有账号？立即注册</Text>
          </TouchableOpacity>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#FFD700',
  },
  scrollContent: {
    flexGrow: 1,
    justifyContent: 'center',
    padding: 20,
  },
  header: {
    alignItems: 'center',
    marginBottom: 40,
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#333',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: '#666',
  },
  form: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 24,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 5,
  },
  input: {
    height: 50,
    borderWidth: 1,
    borderColor: '#ddd',
    borderRadius: 8,
    paddingHorizontal: 16,
    fontSize: 16,
    marginBottom: 16,
  },
  codeRow: {
    flexDirection: 'row',
    marginBottom: 16,
  },
  codeInput: {
    flex: 1,
    marginBottom: 0,
  },
  codeBtn: {
    width: 120,
    height: 50,
    backgroundColor: '#FFD700',
    borderRadius: 8,
    justifyContent: 'center',
    alignItems: 'center',
    marginLeft: 12,
  },
  codeBtnDisabled: {
    backgroundColor: '#ccc',
  },
  codeBtnText: {
    fontSize: 14,
    color: '#333',
    fontWeight: '500',
  },
  switchMode: {
    alignSelf: 'flex-end',
    marginBottom: 20,
  },
  switchModeText: {
    color: '#FFD700',
    fontSize: 14,
  },
  loginBtn: {
    height: 50,
    backgroundColor: '#FFD700',
    borderRadius: 8,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 16,
  },
  loginBtnDisabled: {
    backgroundColor: '#ccc',
  },
  loginBtnText: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#333',
  },
  registerBtn: {
    alignItems: 'center',
  },
  registerBtnText: {
    color: '#666',
    fontSize: 14,
  },
});
