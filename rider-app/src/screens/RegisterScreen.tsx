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

export const RegisterScreen: React.FC = () => {
  const navigation = useNavigation();
  const { setAuth } = useAuthStore();
  
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [code, setCode] = useState('');
  const [nickname, setNickname] = useState('');
  const [inviteCode, setInviteCode] = useState('');
  const [countdown, setCountdown] = useState(0);
  const [loading, setLoading] = useState(false);

  const handleSendCode = async () => {
    if (!phone || phone.length !== 11) {
      Alert.alert('提示', '请输入正确的手机号');
      return;
    }

    try {
      await authApi.sendCode({ phone, type: 'register' });
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

  const handleRegister = async () => {
    if (!phone || phone.length !== 11) {
      Alert.alert('提示', '请输入正确的手机号');
      return;
    }

    if (!code) {
      Alert.alert('提示', '请输入验证码');
      return;
    }

    if (!password || password.length < 6) {
      Alert.alert('提示', '密码至少6位');
      return;
    }

    if (password !== confirmPassword) {
      Alert.alert('提示', '两次密码不一致');
      return;
    }

    if (!nickname) {
      Alert.alert('提示', '请输入昵称');
      return;
    }

    setLoading(true);
    try {
      const result = await authApi.register({
        phone,
        password,
        code,
        nickname,
      });

      await setAuth(result.token, result.refresh_token, result.user_info);
      Alert.alert('成功', '注册成功');
    } catch (error: any) {
      Alert.alert('错误', error.response?.data?.message || '注册失败');
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
          <Text style={styles.title}>注册账号</Text>
          <Text style={styles.subtitle}>成为美团骑手</Text>
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

          <TextInput
            style={styles.input}
            placeholder="请输入密码（至少6位）"
            secureTextEntry
            value={password}
            onChangeText={setPassword}
          />

          <TextInput
            style={styles.input}
            placeholder="请确认密码"
            secureTextEntry
            value={confirmPassword}
            onChangeText={setConfirmPassword}
          />

          <TextInput
            style={styles.input}
            placeholder="请输入昵称"
            value={nickname}
            onChangeText={setNickname}
          />

          <TextInput
            style={styles.input}
            placeholder="邀请码（选填）"
            value={inviteCode}
            onChangeText={setInviteCode}
          />

          <TouchableOpacity
            style={[styles.registerBtn, loading && styles.registerBtnDisabled]}
            onPress={handleRegister}
            disabled={loading}
          >
            <Text style={styles.registerBtnText}>
              {loading ? '注册中...' : '注册'}
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.loginBtn}
            onPress={() => navigation.goBack()}
          >
            <Text style={styles.loginBtnText}>已有账号？立即登录</Text>
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
    marginBottom: 30,
  },
  title: {
    fontSize: 28,
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
  registerBtn: {
    height: 50,
    backgroundColor: '#FFD700',
    borderRadius: 8,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 16,
  },
  registerBtnDisabled: {
    backgroundColor: '#ccc',
  },
  registerBtnText: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#333',
  },
  loginBtn: {
    alignItems: 'center',
  },
  loginBtnText: {
    color: '#666',
    fontSize: 14,
  },
});
