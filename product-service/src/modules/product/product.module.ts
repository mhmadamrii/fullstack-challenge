import { Module } from '@nestjs/common';
import { ProductController } from './product.controller';
import { ProductService } from './product.service';
import { PrismaService } from '../../prisma.service';
import { RedisModule } from '../redis/redis.module';
import { RedisProvider } from '../redis/redis.provider';
import { RabbitMQProvider } from '../rabbitmq/rabbitmq.provider';

@Module({
  imports: [RedisModule],
  controllers: [ProductController],
  providers: [ProductService, PrismaService, RedisProvider, RabbitMQProvider],
})
export class ProductModule {}
