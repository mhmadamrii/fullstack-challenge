import { Module } from '@nestjs/common';
import { OrderController } from './order.controller';
import { OrderService } from './order.service';
import { PrismaService } from 'src/prisma.service';
import { RedisModule } from '../redis/redis.module';
import { RedisProvider } from '../redis/redis.provider';
import { RabbitMQProvider } from '../rabbitmq/rabbitmq.provider';

@Module({
  imports: [RedisModule],
  controllers: [OrderController],
  providers: [OrderService, PrismaService, RedisProvider, RabbitMQProvider],
})
export class OrderModule {}
